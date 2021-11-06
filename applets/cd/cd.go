package cd

import (
	"context"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(CdCommand())
}

func CdCommand() *tish.Command {
	return &tish.Command{
		Name: "cd",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			var dirName string
			if len(env.Args) > 0 {
				dirName = env.Args[0]
			}
			err = env.Shell.SetWorkingDir("cd", dirName, env.Stderr)
			if err != nil {
				result.SetInternalProcessResult(1)
			} else {
				result.SetInternalProcessResult(0)
			}
			return nil
		},
		Completer: nil,
	}
}
