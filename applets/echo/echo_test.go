package echo

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/shibukawa/tish"

	"github.com/stretchr/testify/assert"
)

func Test_echoCommand(t *testing.T) {
	type args struct {
		args []string
	}
	type wants struct {
		stdout string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "single arg, no options",
			args: args{
				args: []string{"hello"},
			},
			wants: wants{
				stdout: "hello\n",
			},
		},
		{
			name: "two Args, no options",
			args: args{
				args: []string{"hello", "world"},
			},
			wants: wants{
				stdout: "hello world\n",
			},
		},
		{
			name: "single arg, no trailing new line options",
			args: args{
				args: []string{"-n", "hello"},
			},
			wants: wants{
				stdout: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/dummy", nil)
			echo := EchoCommand()
			var stdout bytes.Buffer
			p := tish.NewProcess(s, echo.Executor, echo.Name, tt.args.args, 10, 11, nil)
			p.Stdout = &stdout
			p.Stderr = io.Discard
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.stdout, stdout.String())
		})
	}
}
