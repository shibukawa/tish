package ls

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

const binaryName = "ls"

type Directory struct {
	path    string
	Entries []Entry `json:"entries"`
}

const (
	TypeDirectory = "directory"
	TypeRegular   = "regular"
	TypeSymLink   = "symlink"
	TypeHardLink  = "hardlink"
	TypeNamedPipe = "namedpipe"
)

type Entry struct {
	Name       string `json:"name"`
	Mode       string `json:"mode"`
	ModeOctal  int    `json:"mode_octal"`
	User       string `json:"user"`
	Group      string `json:"group"`
	Uid        uint32 `json:"uid"`
	Gid        uint32 `json:"gid"`
	Type       string `json:"type"`
	Size       int64  `json:"size"`
	ModifyTime int64  `json:"modify-time"`
	ModTime    time.Time
}

func init() {
	tish.RegisterCommand(lsCommand())
}

type config struct {
	All           bool `short:"a" description:"all"`
	AllExcludeDot bool `short:"A" description:"all but exclude . and .."`
	FollowLink    bool `short:"L" description:"follow symlink"`
	Long          bool `short:"l" description:"long"`
	Humanize      bool `short:"h" description:"humanize"`
	One           bool `short:"1" description:"one"`
}

func lsCommand() *tish.Command {
	return &tish.Command{
		Name: "ls",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}
			args, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return nil
			}
			err = ls(p.Stdout, args, p.Shell, c)
			if err != nil {
				res.SetInternalProcessResult(1)
				return err
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

func ls(w io.Writer, paths []string, s *tish.Shell, opt *config) error {
	dirs, err := gather(paths, s, opt)
	if err != nil {
		return err
	}

	return output(w, dirs, opt)
}

func gather(paths []string, s *tish.Shell, opt *config) ([]Directory, error) {
	type fileEntry struct {
		fi     os.FileInfo
		orig   string
		direct bool
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}
	ret := make([]Directory, 0)
	for _, path := range paths {
		cur, err := os.Lstat(s.ExpandPath(filepath.Join(path, ".")))
		if err != nil {
			return nil, err
		}
		files := []fileEntry{}
		var orig string
		var direct bool
		if path == "." {
			orig = "."
			direct = false
		} else {
			orig = filepath.Join(path, ".")
			direct = true
		}
		if orig == "." || (!cur.IsDir() && cur.Mode()&os.ModeSymlink == 0) {
			files = append(files, fileEntry{
				fi:     cur,
				orig:   orig,
				direct: direct,
			})
		}
		if path != "/" {
			par, err := os.Lstat(s.ExpandPath(filepath.Join(path, "..")))
			if err != nil {
				return nil, err
			}
			var orig string
			var direct bool
			if path == "." {
				orig = ".."
				direct = false
			} else {
				filepath.Join(path, "..")
				direct = true
			}
			if orig == ".." || !par.IsDir() {
				files = append(files, fileEntry{
					fi:     par,
					orig:   orig,
					direct: direct,
				})
			}
		}
		isDir := cur.IsDir()
		absPath := s.ExpandPath(path)
		if cur.Mode()&os.ModeSymlink != 0 {
			if link, err := os.Readlink(absPath); err == nil {
				if lfi, err :=  os.Lstat(link); err == nil && lfi.IsDir() {
					isDir = true
					absPath = link
				}
			}
			log.Println(isDir, absPath)
		}
		if isDir {
			ff, err := ioutil.ReadDir(absPath)
			if err != nil {
				return nil, err
			}
			for _, f := range ff {
				files = append(files, fileEntry{
					fi:     f,
					orig:   f.Name(),
					direct: false,
				})
			}
		}
		dir := Directory{
			path:    path,
			Entries: make([]Entry, 0, len(files)),
		}
		for _, f := range files {
			if skip(f.orig, f.direct, opt) {
				continue
			}

			var type_ string
			fi := f.fi
			mode := fi.Mode()
			if fi.IsDir() {
				type_ = TypeDirectory
			} else if mode&os.ModeSymlink != 0 {
				type_ = TypeSymLink
				if opt.FollowLink {
					fi, err = followLink(s.ExpandPath(path), fi)
					if err != nil {
						return nil, err
					}
					mode = fi.Mode()
				}
			} else if mode&os.ModeNamedPipe != 0 {
				type_ = TypeNamedPipe
			}

			e := Entry{
				Name:       f.orig,
				Mode:       mode.String(),
				Size:       fi.Size(),
				Type:       type_,
				ModifyTime: fi.ModTime().Unix(),
				ModTime:    fi.ModTime(),
			}

			addUser(&e)

			dir.Entries = append(dir.Entries, e)
		}
		ret = append(ret, dir)
	}

	return ret, nil
}

func skip(fn string, direct bool, opt *config) bool {
	if strings.HasPrefix(fn, ".") {
		if opt.All {
			return false
		}
		if fn == "." || fn == ".." {
			return true
		}
		if opt.AllExcludeDot {
			return false
		}
		return !direct
	}
	return false
}

func followLink(dirPath string, fi os.FileInfo) (os.FileInfo, error) {
	path, err := os.Readlink(filepath.Join(dirPath, fi.Name()))
	if err != nil {
		return nil, err
	}
	return os.Lstat(path)
}
