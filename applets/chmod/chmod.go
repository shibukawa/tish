package chmod

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(chmodCommand())
}

type config struct {
	Recursive bool `short:"R" description:"Change the modes of the file hierarchies rooted in the files, instead of just the files themselves."`
}

func calcMode(info os.FileInfo, plus, mask uint32) os.FileMode {
	mode := uint32(info.Mode())
	remain := mode & (^uint32(0777))
	return os.FileMode((mode&mask)|plus|remain)
}

func chmodCommand() *tish.Command {
	return &tish.Command{
		Name: "chmod",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}
			args, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return err
			}
			if len(args) < 2 {
				res.SetInternalProcessResult(1)
				return errors.New("chmod [mode] ...files")
			}

			plus, mask, err := parseFlag(args[0])
			if err != nil {
				res.SetInternalProcessResult(1)
				return err
			}

			for _, file := range args[1:] {
				path := p.Shell.ExpandPath(file)
				if c.Recursive {
					err := filepath.Walk(path, func(entryPath string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if err := os.Chmod(entryPath, calcMode(info, plus, mask)); err != nil {
							return err
						}
						return nil
					})
					if err != nil {
						res.SetInternalProcessResult(1)
						return err
					}
				} else {
					fi, err := os.Lstat(path)
					if err != nil {
						return err
					}
					os.Chmod(path, calcMode(fi, plus, mask))
				}
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

var octet = regexp.MustCompile(`[0-7]{3}`)

type state int

const (
	target state = iota + 1
	operation
	permission
)

type operationType int

const (
	unset operationType = iota
	add
	minus
	set
)

func parseFlag(input string) (plus, mask uint32, err error) {
	if octet.MatchString(input) {
		v, _ := strconv.ParseUint(input, 8, 32)
		return uint32(v), 0000, nil
	}
	user := false
	group := false
	other := false
	op := unset
	read := false
	write := false
	exec := false

	s := target

	i := 0
	c := input[i]
	exit := false
	for {
		if exit {
			break
		}
		switch s {
		case target:
			switch c {
			case 'u':
				user = true
			case 'g':
				group = true
			case 'o':
				other = true
			case '+', '-', '=':
				s = operation
			default:
				return 0, 0, fmt.Errorf("parse error: need '+-=' and 'rwx': %s", input)
			}
			if s == target {
				i++
				if len(input) == i {
					return 0, 0, fmt.Errorf("parse error: need '+-=' and 'rwx': %s", input)
				}
				c = input[i]
			}
		case operation:
			switch c {
			case '+':
				op = add
			case '-':
				op = minus
			case '=':
				op = set
			default:
				return 0, 0, fmt.Errorf("parse error: need '+-=' and 'rwx': %s", input)
			}
			i++
			if len(input) == i {
				return 0, 0, fmt.Errorf("parse error: need 'rwx': %s", input)
			}
			c = input[i]
			s = permission
		case permission:
			switch c {
			case 'r':
				read = true
			case 'w':
				write = true
			case 'x':
				exec = true
			default:
				return 0, 0, fmt.Errorf("parse error: %s", input)
			}
			i++
			if len(input) == i {
				exit = true
				break
			}
			c = input[i]
		}
	}
	if !user && !group && !other {
		user = true
		group = true
		other = true
	}
	switch op {
	case add:
		var flag uint32 = 00
		var mask uint32 = 07
		if exec {
			flag += 01
			mask -= 01
		}
		if write {
			flag += 02
			mask -= 02
		}
		if read {
			flag += 04
			mask -= 04
		}
		var flagRes uint32 = 0
		var maskRes uint32 = 0777
		if user {
			flagRes |= flag * 0100
			maskRes &= 0077
			maskRes |= mask * 0100
		}
		if group {
			flagRes |= flag * 0010
			maskRes &= 0707
			maskRes |= mask * 0010
		}
		if other {
			flagRes |= flag * 0001
			maskRes &= 0770
			maskRes |= mask * 0001
		}
		return flagRes, maskRes, nil
	case minus:
		var mask uint32 = 07
		if exec {
			mask -= 01
		}
		if write {
			mask -= 02
		}
		if read {
			mask -= 04
		}
		var maskRes uint32 = 0777
		if user {
			maskRes &= 0077
			maskRes |= mask * 0100
		}
		if group {
			maskRes &= 0707
			maskRes |= mask * 0010
		}
		if other {
			maskRes &= 0770
			maskRes |= mask * 0001
		}
		return 0, maskRes, nil
	case set:
		var flag uint32 = 00
		if exec {
			flag += 01
		}
		if write {
			flag += 02
		}
		if read {
			flag += 04
		}
		var flagRes uint32 = 0
		var maskRes uint32 = 0777
		if user {
			flagRes |= flag * 0100
			maskRes &= 0077
		}
		if group {
			flagRes |= flag * 0010
			maskRes &= 0707
		}
		if other {
			flagRes |= flag * 0001
			maskRes &= 0770
		}
		return flagRes, maskRes, nil
	}

	log.Println(user, group, other)

	return 0, 0, nil
}
