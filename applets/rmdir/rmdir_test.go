package rmdir

import (
	"context"
	"log"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_rmdirCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantsErr bool
	}{
		{
			name: "empty dir",
			args: []string{"emptydir"},
		},
		{
			name:     "error: not empty dir",
			args:     []string{"notemptydir"},
			wantsErr: true,
		},
		{
			name:     "error: file",
			args:     []string{"file.txt"},
			wantsErr: true,
		},
		{
			name:     "error: symlink",
			args:     []string{"symlink"},
			wantsErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tish.CreateTestFolders(t, "rmdir", map[string]string{
				"file.txt":             "hello world",
				"emptydir/":            "",
				"notemptydir/test.txt": "test",
				"@symlink":             "dir",
			})

			s := tish.NewShell(root, []string{})
			rd := rmdirCommand()

			p := tish.NewProcess(s, rd.Executor, rd.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())

			if tt.wantsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
