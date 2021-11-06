package mkdir

import (
	"context"
	"os"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(mkdirCommand())
}

type config struct {
	P bool `short:"p" description:"no error if existing, make parent directories as needed"`
}

func mkdirCommand() *tish.Command {
	return &tish.Command{
		Name: "mkdir",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}
			dirs, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return nil
			}

			for _, dir := range dirs {
				dir = p.Shell.ExpandPath(dir)
				if c.P {
					err := os.MkdirAll(dir, 0755)
					if !os.IsExist(err) {
						res.SetInternalProcessResult(1)
						return err
					}
				} else {
					if err := os.Mkdir(dir, 0755); err != nil {
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
