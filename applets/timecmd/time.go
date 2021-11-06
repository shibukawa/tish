package timecmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(TimeCommand())
}

func TimeCommand() *tish.Command {
	return &tish.Command{
		Name: "time",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			if len(env.Args) > 0 {
				cmd := env.Args[0]
				args := env.Args[1:]
				childRes, err := env.Shell.RunChildProcess(ctx, env, cmd, args)
				if childRes.ExitCode() != 0 {
					result.SetInternalProcessResult(childRes.ExitCode())
					return err
				}
				user := float64(childRes.UserTime()) / float64(time.Second)
				system := float64(childRes.SystemTime()) / float64(time.Second)
				wall := float64(childRes.WallTime()) / float64(time.Second)
				cpu := childRes.CPUUsage()

				fmt.Fprintf(env.Stdout, "%s  %.2fs user %.2fs system %d%% cpu %.3f total", strings.Join(env.Args, " "), user, system, cpu, wall)
			}
			return nil
		},
		Completer: nil,
	}
}
