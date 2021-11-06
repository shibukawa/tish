package pushd

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/shibukawa/tish"
)

func init() {
	tish.RegisterCommand(pushdCommand())
	tish.RegisterCommand(popdCommand())
	tish.RegisterCommand(dirsCommand())
}

func pushdCommand() *tish.Command {
	return &tish.Command{
		Name: "pushd",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			current := env.Shell.WorkingDir()
			var dirName string
			if len(env.Args) > 0 {
				dirName = env.Args[0]
			}
			err = env.Shell.SetWorkingDir("pushd", dirName, env.Stderr)
			if err != nil {
				result.SetInternalProcessResult(1)
				return err
			}
			env.Shell.Dirs = append(env.Shell.Dirs, current)
			showDirStack(env.Shell, env.Stdout)
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

func popdCommand() *tish.Command {
	return &tish.Command{
		Name: "popd",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			if len(env.Shell.Dirs) == 0 {
				io.WriteString(env.Stderr, "popd: directory stack empty\n")
				result.SetInternalProcessResult(1)
				return tish.ErrStackEmpty
			}
			last := env.Shell.Dirs[len(env.Shell.Dirs)-1]
			env.Shell.Dirs = env.Shell.Dirs[:len(env.Shell.Dirs)-1]
			env.Shell.SetWorkingDir("popd", last, env.Stderr)
			showDirStack(env.Shell, env.Stdout)
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

func normalizeDir(home, target string) string {
	if home == target {
		return "~"
	}
	rel, err := filepath.Rel(home, target)
	if err != nil || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return target
	}
	return filepath.Join("~", rel)
}

func dirsCommand() *tish.Command {
	return &tish.Command{
		Name: "dirs",
		Executor: func(ctx context.Context, result *tish.ExecResult, env *tish.Process) (err error) {
			showDirStack(env.Shell, env.Stdout)
			result.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

func showDirStack(s *tish.Shell, stdout io.Writer) {
	home, _ := s.HomeDir()
	io.WriteString(stdout, normalizeDir(home, s.WorkingDir()))
	for i := len(s.Dirs) - 1; i >= 0; i-- {
		io.WriteString(stdout, " "+normalizeDir(home, s.Dirs[i]))
	}
	io.WriteString(stdout, "\n")
}
