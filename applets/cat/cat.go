package cat

import (
	"context"
	"io"
	"os"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(CatCommand())
}

func CatCommand() *tish.Command {
	return &tish.Command{
		Name: "cat",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			for _, path := range p.Args {
				f, err := os.Open(p.Shell.ExpandPath(path))
				if err != nil {
					return err
				}
				defer f.Close()
				io.Copy(p.Stdout, f)
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

