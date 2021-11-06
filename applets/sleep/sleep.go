package sleep

import (
	"context"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/shibukawa/tish"
)

var parseNum = regexp.MustCompile(`([\d.]+)`)

func init() {
	tish.RegisterCommand(SleepCommand())
}

func SleepCommand() *tish.Command {
	return &tish.Command{
		Name: "sleep",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			if len(env.Args) == 0 {
				io.WriteString(env.Stderr, "usage: sleep seconds\n")
				result.SetInternalProcessResult(1)
				return tish.ErrRequireParameter
			}
			var duration time.Duration
			m := parseNum.FindStringSubmatch(env.Args[0])
			if len(m) > 0 {
				num, err := strconv.ParseFloat(m[1], 64)
				if err == nil {
					duration = time.Duration(num * float64(time.Second))
				}
			}
			time.Sleep(duration)
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}
