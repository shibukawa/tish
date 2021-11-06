package cp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/shibukawa/tish"
	"github.com/shibukawa/tish/osutils"
	"github.com/jessevdk/go-flags"
)

type config struct {
	Force         bool `short:"f" description:"Don't prompt before overwriting"`
	Recursive     bool `short:"R" alias:"r" description:"Copy file hierarchies"`
	FollowSymLink bool `short:"L" description:"always follow symbolic links in SOURCE"`
	Preserve      bool `short:"p" description:"Duplicate the following characteristics of each source file in the corresponding destination file"`
	Update        bool `short:"u" description:"copy only when the SOURCE file is newer than the destination file"`
}

func init() {
	tish.RegisterCommand(CpCommand())
}

func CpCommand() *tish.Command {
	return &tish.Command{
		Name: "cp",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}

			args, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return err
			}
			err = cp(args, p.Shell, c)
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

func cp(files []string, s *tish.Shell, c *config) error {
	dest := s.ExpandPath(files[len(files)-1])

	destIsDir := osutils.IsDir(dest)

	for _, src := range files[:len(files)-1] {
		var dstPath string
		if destIsDir {
			dstPath = filepath.Join(dest, filepath.Base(src))
		} else {
			dstPath = dest
		}
		if !c.Force && osutils.FileExists(dstPath) {
			return os.ErrExist
		}
		srcPath := s.ExpandPath(src)
		if err := cpOne(srcPath, dstPath, s, c); err != nil {
			return err
		}
	}
	return nil
}

func cpOne(srcPath, dstPath string, s *tish.Shell, c *config) error {
	si, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	if si.IsDir() {
		if !c.Recursive {
			return errors.New("source is a directory")
		}
		if err := cpDirectory(srcPath, dstPath, s, c); err != nil {
			return err
		}
	} else {
		if err := cpFile(srcPath, dstPath, s, c); err != nil {
			return err
		}
	}

	return nil
}

func cpDirectory(srcPath, dstPath string, s *tish.Shell, c *config) error {
	si, err := os.Stat(srcPath)
	if err != nil {
		return err
	}

	// ensure dst dir does not already exist
	if _, err := os.Open(dstPath); !os.IsNotExist(err) {
		return errors.New("destination already exists")
	}

	// create dst dir
	if err := os.MkdirAll(dstPath, si.Mode()); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := cpOne(filepath.Join(srcPath, file.Name()),
			filepath.Join(dstPath, file.Name()), s, c); err != nil {
			return err
		}
	}
	return nil
}

func cpFile(srcPath, dstPath string, s *tish.Shell, c *config) error {
	si, err := os.Lstat(srcPath)
	if err != nil {
		return err
	}

	if c.Preserve && !si.Mode().IsRegular() {
		return cpSymlink(srcPath, dstPath, s)
	}

	di, err := os.Lstat(dstPath)
	if !os.IsNotExist(err) {
		if !c.Force {
			return fmt.Errorf("destination already exists: %s", dstPath)
		}
		if si.ModTime().After(di.ModTime()) && c.Update {
			return errors.New("destination is newer then src")
		}
	}

	//open source
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	//create dst
	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	// copy
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	if err = out.Chmod(si.Mode()); err != nil {
		return err
	}
	if c.Preserve {
		if err = os.Chtimes(dstPath, si.ModTime(), si.ModTime()); err != nil {
			return err
		}
	}

	//sync dst to disk
	return out.Sync()
}

func cpSymlink(src, dst string, s *tish.Shell) error {
	linkTarget, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(linkTarget, dst)
}
