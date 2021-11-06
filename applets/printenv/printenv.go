package printenv

import (
	"context"
	"fmt"
	"sort"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(PrintenvCommand())
}

func PrintenvCommand() *tish.Command {
	return &tish.Command{
		Name: "printenv",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			e := env.Env()
			if len(env.Args) > 0 {
				fmt.Fprintln(env.Stdout, e[env.Args[0]])
			} else {
				var keys []string
				for key := range e {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					fmt.Fprintf(env.Stdout, "%s=%s\n", key, e[key])
				}
			}
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}
