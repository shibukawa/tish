package tish

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/shibukawa/tish/parser"
)

var (
	ErrExit                 = errors.New("exit")
	ErrNoPipeReceiver       = errors.New("no pipe receiver")
	ErrCommandNotFound      = errors.New("command not found")
	ErrCommandError         = errors.New("command error")
	ErrRedirectError        = errors.New("redirect error")
	ErrWildcardNoMatchError = errors.New("wild card no match")
)

type ErrCmdNotFound struct {
	Command     string
	ParsedArgs  []string
	FullCommand string
}

func (e ErrCmdNotFound) Error() string {
	return fmt.Sprintf("command not found: %s", e.Command)
}

type ErrFileNotFound struct {
	Command  string
	NotFound string
	Current  string
}

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("%s no such file or directory: %s", e.Command, e.NotFound)
}

var EnvVarPattern = regexp.MustCompile(`([a-zA-Z_]+[a-zA-Z0-9_]*)=(.*)`)

type Option struct {
	IgnoreError bool     `json:"ignore_error"`
	AllowPath   []string `json:"allow_path"`
}

var (
	pidLock = &sync.Mutex{}
	lastPid = 10
)

func newProcessID() int {
	pidLock.Lock()
	defer pidLock.Unlock()
	lastPid++
	return lastPid
}

type Shell struct {
	wd       string
	commands []*Command
	Env      map[string]string
	Dirs     []string
	Pid      int
	lock     *sync.Mutex
	option   Option
}

type CurrentShellStatus struct {
	s    *Shell
	envs map[string]string
	pid  int
}

func NewShell(cwd string, envs []string, opt ...Option) *Shell {
	envMap := map[string]string{}
	for _, env := range envs {
		m := EnvVarPattern.FindStringSubmatch(env)
		if len(m) == 0 {
			envMap[env] = ""
		} else {
			envMap[m[1]] = m[2]
		}
	}
	s := &Shell{
		wd:       cwd,
		Env:      envMap,
		lock:     &sync.Mutex{},
		commands: commands,
		Pid:      newProcessID(),
	}
	if len(opt) > 0 {
		s.option = opt[0]
	}
	return s
}

func (s *Shell) Run(ctx context.Context, cmdStr string, stdout, stderr io.Writer) (code int, err error) {
	sessionGroups, err := parser.ParseCommandStr(cmdStr)
	if err != nil {
		return 1, err
	}
	s.runSessionGroups(ctx, sessionGroups, stdout, stderr)
	return 0, nil
}

func (s *Shell) runSessionGroups(ctx context.Context, sessionGroups [][]*parser.Session, stdout, stderr io.Writer) (result *ExecResult, err error) {
	for _, sg := range sessionGroups {
		result, err = s.runSessionGroup(ctx, sg, stdout, stderr)
		last := sg[len(sg)-1]
		switch last.Separator {
		case parser.Semicolon:
			// do nothing
		case parser.LogicalOr:
			if result.ExitCode() == 0 {
				return
			}
		case parser.LogicalAnd:
			if result.ExitCode() != 0 {
				return
			}
		}
	}
	return
}

func (s *Shell) runSessionGroup(ctx context.Context, sessions []*parser.Session, stdout, stderr io.Writer) (*ExecResult, error) {
	var procs []*Process
	for i, ses := range sessions {
		pid := newProcessID()
		for _, f := range ses.Fragments {
			for _, subSes := range f.Sessions {
				cmdName, args := subSes.GetCommand()
				cmd := s.lookupCommand(cmdName)
				if cmd == nil {
					return nil, fmt.Errorf("command '%s' is not found: %w", cmdName, ErrCmdNotFound{})
				}
				subProc := NewProcess(s, cmd.Executor, cmdName, args, pid, newProcessID(), s.Env)
				var stdout bytes.Buffer
				subProc.Stdout = &stdout
				err := subProc.StartAndWait(ctx)
				if err != nil {
					// todo: human readable error
					return nil, ErrCommandError
				}
				f.Term = stdout.String()
			}
		}
		cmdName, args := ses.GetCommand()
		cmd := s.lookupCommand(cmdName)
		if cmd == nil {
			return nil, fmt.Errorf("command '%s' is not found: %w", cmdName, ErrCmdNotFound{})
		}
		proc := NewProcess(s, cmd.Executor, cmdName, args, s.Pid, pid, s.Env)
		proc.Stdout = stdout
		proc.Stderr = stderr
		if i != 0 {
			procs[i-1].Pipe(proc)
		}
		if ses.Stdin != "" {
			if i != 0 {
				return nil, ErrRedirectError
			}
			err := proc.RedirectStdin(ses.Stdin)
			if err != nil {
				return nil, err
			}
		}
		if ses.Stdout != "" {
			if i != len(sessions)-1 {
				return nil, ErrRedirectError
			}
			err := proc.RedirectStdout(ses.Stdout, ses.StdoutAppend)
			if err != nil {
				return nil, err
			}
		}
		if ses.Stderr != "" {
			if i != len(sessions)-1 {
				return nil, ErrRedirectError
			}
			err := proc.RedirectStderr(ses.Stderr, ses.StderrAppend)
			if err != nil {
				return nil, err
			}
		}
		procs = append(procs, proc)
	}
	for _, proc := range procs {
		err := proc.Start(ctx)
		if err != nil {
			// todo: human readable error
			return nil, err
		}
	}
	for _, proc := range procs {
		proc.Wait()
	}
	return procs[len(procs)-1].Result, nil
}

func (s *Shell) RunChildProcess(ctx context.Context, p *Process, cmdName string, args []string) (*ExecResult, error) {
	cmd := s.lookupCommand(cmdName)
	if cmd == nil {
		return nil, fmt.Errorf("command '%s' is not found: %w", cmdName, ErrCmdNotFound{})
	}
	proc := NewProcess(s, cmd.Executor, cmdName, args, p.Pid, newProcessID(), p.Env())
	err := proc.StartAndWait(ctx)
	return proc.Result, err
}

func (s *Shell) SetEnv(key, value string) {
	s.Env[key] = value
}

func (s *Shell) DelEnv(key string) {
	delete(s.Env, key)
}

func (s *Shell) lookupCommand(cmdName string) *Command {
	// todo: internal cmmand only mode (safe mode)
	for _, cmd := range s.commands {
		if cmd.Name == cmdName {
			return cmd
		}
	}
	if cmd, err := lookupExternalCommand(cmdName); err == nil {
		return cmd
	}
	return nil
}

func (s Shell) WorkingDir() string {
	return s.wd
}

func (s *Shell) SetWorkingDir(commandName, dirName string, stderr io.Writer) error {
	var err error
	if dirName == "" {
		dirName, err = s.HomeDir()
		if err != nil {
			return err
		}
	}
	wd := s.ExpandPath(dirName)
	_, err = os.Lstat(wd)
	if os.IsNotExist(err) {
		if stderr != nil {
			fmt.Fprintf(stderr, "%s: no such file or directory: %s\n", commandName, dirName)
		}
		return &ErrFileNotFound{
			Command:  commandName,
			NotFound: dirName,
			Current:  s.WorkingDir(),
		}
	}
	s.wd = wd
	return nil
}

func (s *Shell) HomeDir() (string, error) {
	env, enverr := "HOME", "$HOME"
	switch runtime.GOOS {
	case "windows":
		env, enverr = "USERPROFILE", "%userprofile%"
	case "plan9":
		env, enverr = "home", "$home"
	}
	if v := s.Env[env]; v != "" {
		return v, nil
	}
	switch runtime.GOOS {
	case "android":
		return "/sdcard", nil
	case "ios":
		return "/", nil
	}
	return "", errors.New(enverr + " is not defined")
}

func (s *Shell) expandWildcard(f parser.Fragment) ([]string, error) {
	if !strings.ContainsAny(f.Term, "*?[") {
		return []string{f.Term}, nil
	}
	path := f.Term
	files, err := filepath.Glob(s.ExpandPath(path))
	if len(files) == 0 {
		return nil, ErrWildcardNoMatchError
	}
	return files, err
}

func (s Shell) ExpandPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(s.WorkingDir(), path)
}