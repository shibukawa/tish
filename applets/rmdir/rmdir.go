package rmdir

import (
	"context"
	"fmt"
	"os"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(rmdirCommand())
}

func rmdirCommand() *tish.Command {
	return &tish.Command{
		Name: "rmdir",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			for _, dir := range p.Args {
				dirPath := p.Shell.ExpandPath(dir)
				fi, err := os.Lstat(dirPath)
				if err != nil {
					res.SetInternalProcessResult(1)
					return err
				}
				if !fi.IsDir() {
					res.SetInternalProcessResult(1)
					return fmt.Errorf("rmdir: %s: Not a directory", dir)
				}
				if err = os.Remove(dirPath); err != nil {
					res.SetInternalProcessResult(1)
					return err
				}
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}
