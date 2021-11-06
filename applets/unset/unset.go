package unset

import (
	"context"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(UnsetCommand())
}

func UnsetCommand() *tish.Command {
	return &tish.Command{
		Name: "unset",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			conf := &struct {
				Variable bool `short:"v"`
			}{}
			args, err := flags.ParseArgs(conf, env.Args)
			if err != nil {
				result.SetInternalProcessResult(1)
				return err
			}
			for _, key := range args {
				delete(env.Shell.Env, key)
			}
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}
