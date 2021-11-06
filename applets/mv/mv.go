package mv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shibukawa/tish"
	"github.com/shibukawa/tish/osutils"
)

func init() {
	tish.RegisterCommand(MvCommand())
}

func MvCommand() *tish.Command {
	return &tish.Command{
		Name: "mv",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			if len(p.Args) < 2 {
				res.SetInternalProcessResult(1)
				return errors.New("usage: mv source ... directory")
			}
			dest := p.Shell.ExpandPath(p.Args[len(p.Args)-1])
			destIsDir := osutils.IsDir(dest)
			if len(p.Args) == 2 {
				src := p.Args[0]
				var destPath string
				if destIsDir {
					destPath = filepath.Join(dest, filepath.Base(src))
				} else {
					destPath = dest
				}
				err := os.Rename(p.Shell.ExpandPath(src), destPath)
				if err != nil {
					res.SetInternalProcessResult(1)
					return err
				}
			} else {
				if !destIsDir {
					res.SetInternalProcessResult(1)
					return fmt.Errorf("destination '%s' should be directory", p.Args[len(p.Args)-1])
				}
				for _, src := range p.Args[:len(p.Args)-1] {
					destPath := filepath.Join(dest, filepath.Base(src))
					err := os.Rename(p.Shell.ExpandPath(src), destPath)
					if err != nil {
						res.SetInternalProcessResult(1)
						return err
					}
				}
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

