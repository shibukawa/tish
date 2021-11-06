package cat

import (
	"bytes"
	"context"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func TestCatCommand(t *testing.T) {
	root := tish.CreateTestFolders(t, "wc", map[string]string{
		"hello.txt": "hello\n",
		"world.txt": "world\n",
	})

	tests := []struct {
		name string
		args []string
		wants string
	}{
		{
			name: "single file",
			args: []string{"hello.txt"},
			wants: "hello\n",
		},
		{
			name: "multiple files",
			args: []string{"hello.txt", "world.txt"},
			wants: "hello\nworld\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(root, []string{})
			c := CatCommand()
			buf := bytes.Buffer{}
			p := tish.NewProcess(s, c.Executor, c.Name, tt.args, 10, 11, nil)
			p.Stdout = &buf
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, 0, p.Result.ExitCode())
			assert.Equal(t, tt.wants, buf.String())
		})
	}
}