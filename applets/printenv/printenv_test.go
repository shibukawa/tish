package printenv

import (
	"bytes"
	"context"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func Test_printenvCommand(t *testing.T) {
	type args struct {
		envs  []string
		param []string
	}
	type wants struct {
		envs   map[string]string
		stdout string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "print all",
			args: args{
				envs:  []string{"A=B", "C=D"},
				param: []string{},
			},
			wants: wants{
				stdout: `A=B
C=D
`,
			},
		},
		{
			name: "print single variable",
			args: args{
				envs:  []string{"A=B", "C=D"},
				param: []string{"A"},
			},
			wants: wants{
				stdout: "B\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/home/myname", tt.args.envs)
			e := PrintenvCommand()
			buf := bytes.Buffer{}
			p := tish.NewProcess(s, e.Executor, e.Name, tt.args.param, 10, 11, nil)
			p.Stdout = &buf
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.stdout, buf.String())
		})
	}
}
