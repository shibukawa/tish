package cd

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/shibukawa/tish"

	"github.com/stretchr/testify/assert"
)

func Test_cdCommand(t *testing.T) {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("tish-cd-test-%d", rand.Int()))
	os.MkdirAll(filepath.Join(root, "current", "sub1", "sub2"), 0o777)
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	type args struct {
		wd    string
		envs  []string
		param []string
	}
	type wants struct {
		wd     string
		status int
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "relative",
			args: args{
				wd:    filepath.Join(root, "current"),
				envs:  []string{},
				param: []string{".."},
			},
			wants: wants{
				wd:     root,
				status: 0,
			},
		},
		{
			name: "absolute",
			args: args{
				wd:    filepath.Join(root, "current"),
				envs:  []string{},
				param: []string{root},
			},
			wants: wants{
				wd:     root,
				status: 0,
			},
		},
		{
			name: "not found",
			args: args{
				wd:    filepath.Join(root, "current"),
				envs:  []string{},
				param: []string{"not-exist"},
			},
			wants: wants{
				wd:     filepath.Join(root, "current"),
				status: 1,
			},
		},
		{
			name: "home",
			args: args{
				wd:    filepath.Join(root, "current"),
				envs:  []string{"HOME=" + root},
				param: []string{},
			},
			wants: wants{
				wd:     root,
				status: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(tt.args.wd, tt.args.envs)
			cmd := CdCommand()
			stderr := bytes.Buffer{}
			p := tish.NewProcess(s, cmd.Executor, cmd.Name, tt.args.param,11, 10, nil)
			p.Stderr = &stderr
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.wd, s.WorkingDir())
			assert.Equal(t, tt.wants.status, p.Result.ExitCode())
		})
	}
}
