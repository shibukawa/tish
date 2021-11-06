package export

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(ExportCommand())
}

var envVarPattern = regexp.MustCompile(`([a-zA-Z_]+[a-zA-Z0-9_]*)=(.*)`)

func ExportCommand() *tish.Command {
	return &tish.Command{
		Name: "export",
		Executor: func(ctx context.Context, res *tish.ExecResult, env *tish.Process) (err error) {
			conf := &struct {
				Delete bool `short:"n"`
				Print  bool `short:"p"`
			}{}
			args, err := flags.ParseArgs(conf, env.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				log.Println(err)
				return nil
			}
			if conf.Delete {
				for _, key := range args {
					env.Shell.DelEnv(key)
				}
			} else if conf.Print {
				var keys []string
				e := env.Env()
				for key := range e {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					fmt.Fprintf(env.Stdout, "declare -x %s=\"%s\"\n", key, e[key])
				}
			} else {
				for _, arg := range args {
					m := envVarPattern.FindStringSubmatch(arg)
					if len(m) == 0 { // VAR_NAME
						env.Shell.SetEnv(arg, "")
					} else {
						env.Shell.SetEnv(m[1], m[2])
					}
				}
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}
