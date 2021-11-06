package export

import (
	"bytes"
	"context"
	"testing"

	"github.com/shibukawa/tish"

	"github.com/stretchr/testify/assert"
)

func Test_exportCommand(t *testing.T) {
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
			name: "key define",
			args: args{
				envs:  nil,
				param: []string{"A"},
			},
			wants: wants{
				envs: map[string]string{
					"A": "",
				},
				stdout: "",
			},
		},
		{
			name: "value define",
			args: args{
				envs:  nil,
				param: []string{"A=B"},
			},
			wants: wants{
				envs: map[string]string{
					"A": "B",
				},
				stdout: "",
			},
		},
		{
			name: "delete (success)",
			args: args{
				envs:  []string{"A=B"},
				param: []string{"-n", "A"},
			},
			wants: wants{
				envs:   map[string]string{},
				stdout: "",
			},
		},
		{
			name: "delete (missing)",
			args: args{
				envs:  []string{"A=B"},
				param: []string{"-n", "Z"},
			},
			wants: wants{
				envs: map[string]string{
					"A": "B",
				},
				stdout: "",
			},
		},
		{
			name: "print",
			args: args{
				envs:  []string{"A=B", "C=D"},
				param: []string{"-p"},
			},
			wants: wants{
				envs: map[string]string{
					"A": "B",
					"C": "D",
				},
				stdout: `declare -x A="B"
declare -x C="D"
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/home/myname", tt.args.envs)
			e := ExportCommand()
			buf := bytes.Buffer{}

			p := tish.NewProcess(s, e.Executor, e.Name, tt.args.param, 10, 11, nil)
			p.Stdout = &buf
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.envs, s.Env)
			assert.Equal(t, tt.wants.stdout, buf.String())
		})
	}
}
