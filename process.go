package tish

import (
	"context"
	"io"
	"os"
	"sync"
	"time"
)

type Executor func(ctx context.Context, result *ExecResult, p *Process) (err error)

type Completer func(input string) []string

type Process struct {
	Shell *Shell

	ParentPid int
	Pid       int

	env  map[string]string
	Cmd  string
	Args []string
	OrigArgs []string

	Stdin        io.Reader
	StdinCloser  io.Closer
	Stdout       io.Writer
	StdoutCloser io.Closer
	Stderr       io.Writer
	StderrCloser io.Closer

	Executor Executor

	Result *ExecResult

	wg        *sync.WaitGroup
	execError error
}

func NewProcess(s *Shell, executor Executor, cmd string, args []string, ppid, pid int, env map[string]string) *Process {
	if env == nil {
		env = s.Env
	}
	return &Process{
		Shell:     s,
		Pid:       pid,
		ParentPid: ppid,
		Cmd:       cmd,
		OrigArgs:      args,
		env:       env,
		Executor:  executor,
		wg:        &sync.WaitGroup{},
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}
}

func (p *Process) Pipe(other *Process) {
	reader, writer := io.Pipe()
	p.Stdout = writer
	p.StdoutCloser = writer
	other.Stdin = reader
	other.StdinCloser = reader
}

func (p *Process) RedirectStdin(path string) error {
	f, err := os.Open(p.Shell.ExpandPath(path))
	if err != nil {
		return err
	}
	p.Stdin = f
	p.StdinCloser = f
	return nil
}

func (p *Process) RedirectStdout(path string, append bool) error {
	flag := os.O_CREATE | os.O_WRONLY
	if append {
		flag += os.O_APPEND
	} else {
		flag += os.O_TRUNC
	}
	f, err := os.OpenFile(p.Shell.ExpandPath(path), flag, 0o777)
	if err != nil {
		return err
	}
	p.Stdout = f
	p.StdoutCloser = f
	return nil
}

func (p *Process) RedirectStderr(path string, append bool) error {
	flag := os.O_CREATE | os.O_WRONLY
	if append {
		flag += os.O_APPEND
	} else {
		flag += os.O_TRUNC
	}
	f, err := os.OpenFile(p.Shell.ExpandPath(path), flag, 0o777)
	if err != nil {
		return err
	}
	p.Stderr = f
	p.StderrCloser = f
	return nil
}

func (e Process) Env() map[string]string {
	result := map[string]string{}
	for k, v := range e.Shell.Env {
		result[k] = v
	}
	for k, v := range e.env {
		result[k] = v
	}
	return result
}

func (p *Process) StartAndWait(ctx context.Context) error {
	err := p.Start(ctx)
	if err != nil {
		return err
	}
	return p.Wait()
}

func (p *Process) Start(ctx context.Context) error {
	r, err := newExecResult()
	if err != nil {
		return err
	}
	p.Result = r
	env := p.Env()
	for _, arg := range p.OrigArgs {
		p.Args = append(p.Args, os.Expand(arg, func(key string) string {
			return env[key]
		}))
	}
	p.wg.Add(1)
	go func() {
		p.execError = p.Executor(ctx, r, p)
		r.Finish()
		p.wg.Done()
	}()
	return nil
}

func (p *Process) Wait() error {
	p.wg.Wait()
	if p.StdinCloser != nil {
		p.StdinCloser.Close()
	}
	if p.StdoutCloser != nil {
		p.StdoutCloser.Close()
	}
	if p.StderrCloser != nil {
		p.StderrCloser.Close()
	}
	return p.execError
}

type ExecResult struct {
	start time.Time
	// initialStat cpu.TimesStat
	state    *os.ProcessState
	system   time.Duration
	user     time.Duration
	wall     time.Duration
	exitCode int
}

func newExecResult() (*ExecResult, error) {
	/*
		s, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
	*/
	return &ExecResult{
		start: time.Now(),
		// initialStat: s[0],
	}, nil
}

func (e *ExecResult) SetExternalProcessResult(state *os.ProcessState) error {
	e.state = state
	return nil
}

func (e *ExecResult) Finish() {
	e.wall = time.Now().Sub(e.start)
}

func (e ExecResult) WallTime() time.Duration {
	return e.wall
}

func (e ExecResult) SystemTime() time.Duration {
	if e.state != nil {
		e.state.SystemTime()
	}
	return time.Duration(0)
}

func (e ExecResult) UserTime() time.Duration {
	if e.state != nil {
		e.state.UserTime()
	}
	return time.Duration(0)
}

func (e ExecResult) CPUUsage() int {
	return 0
}

func (e *ExecResult) SetInternalProcessResult(exitCode int) error {
	/* s, err := cpu.Times(false)
	if err != nil {
		return err
	}
	*/
	e.exitCode = exitCode
	// e.system = time.Duration(s[0].System) * time.Second
	// e.user = time.Duration(s[0].User) * time.Second
	return nil
}

func (e ExecResult) ExitCode() int {
	if e.state != nil {
		return e.state.ExitCode()
	}
	return e.exitCode
}

type Command struct {
	Name      string
	Executor  Executor
	Completer Completer
}
