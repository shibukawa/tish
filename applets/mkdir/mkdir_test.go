package mkdir

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

func Test_mkdirCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantsErr bool
	}{
		{
			name: "no -p: create new dir",
			args: []string{"newdir"},
		},
		{
			name: "no -p: create new dir under existing dir",
			args: []string{"dir/newdir"},
		},
		{
			name: "no -p: create new dir under existing dir(symlink)",
			args: []string{"symlink/newdir"},
		},
		{
			name:     "error: no -p: create new dirs",
			args:     []string{"parent/newdir"},
			wantsErr: true,
		},
		{
			name:     "error: create new dir as same as existing dir",
			args:     []string{"dir"},
			wantsErr: true,
		},
		{
			name:     "error: create new dir as same as existing file",
			args:     []string{"file.txt"},
			wantsErr: true,
		},
		{
			name: "with -p: create new dirs",
			args: []string{"-p", "parent/newdir"},
		},
		{
			name: "with -p: create new dirs under existing dir",
			args: []string{"-p", "dir/parent/newdir"},
		},
		{
			name: "with -p: create new dirs under existing dir(symlink)",
			args: []string{"-p", "symlink/parent/newdir"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tish.CreateTestFolders(t, "mkdir", map[string]string{
				"file.txt": "hello world",
				"dir/":     "",
				"@symlink": "dir",
			})

			s := tish.NewShell(root, []string{})
			md := mkdirCommand()

			p := tish.NewProcess(s, md.Executor, md.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())

			if tt.wantsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
