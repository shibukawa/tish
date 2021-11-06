package tish

import (
	"context"
	"os/exec"
)

type externalCommand struct {
	fullPath string
}

func (e externalCommand) Executor(ctx context.Context, result *ExecResult, p *Process) (err error) {
	cmd := exec.CommandContext(ctx, e.fullPath, p.Args...)
	cmd.Stdin = p.Stdin
	cmd.Stdout = p.Stdout
	cmd.Stderr = p.Stderr
	cmd.Dir = p.Shell.WorkingDir()
	// todo: env
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	p.Result.SetExternalProcessResult(cmd.ProcessState)
	return err
}

func lookupExternalCommand(cmd string) (*Command, error) {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return nil, err
	}
	ec := &externalCommand{
		fullPath: path,
	}
	return &Command{
		Name:      "cmd",
		Executor:  ec.Executor,
		Completer: nil,
	}, nil
}
