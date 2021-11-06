package unset

import (
	"bytes"
	"context"
	"testing"

	"github.com/shibukawa/tish"

	"github.com/stretchr/testify/assert"
)

func Test_unsetCommand(t *testing.T) {
	type args struct {
		envs  []string
		param []string
	}
	type wants struct {
		envs map[string]string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "delete",
			args: args{
				envs:  []string{"A=B"},
				param: []string{"A"},
			},
			wants: wants{
				envs: map[string]string{},
			},
		},
		{
			name: "delete (with -v)",
			args: args{
				envs:  []string{"A=B"},
				param: []string{"-v", "A"},
			},
			wants: wants{
				envs: map[string]string{},
			},
		},
		{
			name: "delete (missing)",
			args: args{
				envs:  []string{"A=B"},
				param: []string{"B"},
			},
			wants: wants{
				envs: map[string]string{
					"A": "B",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/home/myname", tt.args.envs)
			u := UnsetCommand()
			buf := bytes.Buffer{}
			p := tish.NewProcess(s, u.Executor, u.Name, tt.args.param,  10, 11, nil)
			p.Stdout = &buf
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.envs, s.Env)
		})
	}
}
