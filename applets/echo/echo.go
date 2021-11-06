package echo

import (
	"context"
	"io"
	"log"

	"github.com/shibukawa/tish"

	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(EchoCommand())
}

func EchoCommand() *tish.Command {
	return &tish.Command{
		Name: "echo",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			conf := &struct {
				NoTrailingNewLine bool `short:"n"`
			}{}
			args, err := flags.ParseArgs(conf, env.Args)
			if err != nil {
				log.Println(err)
				result.SetInternalProcessResult(1)
				return err
			}
			for i, arg := range args {
				if i != 0 {
					io.WriteString(env.Stdout, " ")
				}
				io.WriteString(env.Stdout, arg)
			}
			if !conf.NoTrailingNewLine {
				io.WriteString(env.Stdout, "\n")
			}
			result.SetInternalProcessResult(0)
			return err
		},
		Completer: nil,
	}
}
