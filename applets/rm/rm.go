package rm

import (
	"context"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(rmCommand())
}

type config struct {
	Recursive      bool `short:"R" description:"Attempt to remove the file hierarchy rooted in each file argument."`
	RecursiveAlias bool `short:"r" description:"Equivalent to -R."`
	Force          bool `short:"f" description:"ttempt to remove the files without prompting for confirmation."`
}

func rmCommand() *tish.Command {
	return &tish.Command{
		Name: "rm",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}
			dirs, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return nil
			}

			for _, dir := range dirs {
				dirPath := p.Shell.ExpandPath(dir)
				var err error
				if c.Recursive || c.RecursiveAlias {
					err = os.RemoveAll(dirPath)
				} else {
					fi, err := os.Lstat(dirPath)
					if err != nil {
						res.SetInternalProcessResult(1)
						return err
					}
					if fi.IsDir() {
						res.SetInternalProcessResult(1)
						return fmt.Errorf("rm: %s: is a directory", dir)
					}
					err = os.Remove(dirPath)
				}
				if err != nil {
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
