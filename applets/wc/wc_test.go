package wc

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

// 4 lines, 4 words, 25 bytes
var testContent = `hello
world

good night
`

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_wcCommand(t *testing.T) {
	root := tish.CreateTestFolders(t, "wc", map[string]string{
		"test.txt": testContent,
	})

	type args struct {
		param []string
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
			name: "file: lines",
			args: args{
				param: []string{"-l", "test.txt"},
			},
			wants: wants{
				stdout: "4 test.txt\n",
			},
		},
		{
			name: "file: words",
			args: args{
				param: []string{"-w", "test.txt"},
			},
			wants: wants{
				stdout: "4 test.txt\n",
			},
		},
		{
			name: "file: chars",
			args: args{
				param: []string{"-c", "test.txt"},
			},
			wants: wants{
				stdout: "24 test.txt\n",
			},
		},
		{
			name: "file: all",
			args: args{
				param: []string{"test.txt"},
			},
			wants: wants{
				stdout: "4 4 24 test.txt\n",
			},
		},
		{
			name: "stdin: lines",
			args: args{
				param: []string{"-l"},
			},
			wants: wants{
				stdout: "4\n",
			},
		},
		{
			name: "stdin: words",
			args: args{
				param: []string{"-w", "-"},
			},
			wants: wants{
				stdout: "4\n",
			},
		},
		{
			name: "stdin: chars",
			args: args{
				param: []string{"-c"},
			},
			wants: wants{
				stdout: "24\n",
			},
		},
		{
			name: "stdin: all",
			args: args{
				param: []string{"-"},
			},
			wants: wants{
				stdout: "4 4 24\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(root, []string{})
			e := WordCountCommand()
			buf := bytes.Buffer{}

			p := tish.NewProcess(s, e.Executor, e.Name, tt.args.param, 10, 11, nil)
			p.Stdin = strings.NewReader(testContent)
			p.Stdout = &buf
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, 0, p.Result.ExitCode())
			assert.Equal(t, tt.wants.stdout, buf.String())
		})
	}
}
