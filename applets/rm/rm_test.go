package rm

import (
	"context"
	"log"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/shibukawa/tish/osutils"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_rmCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		notExists []string
		wantsErr  bool
	}{
		{
			name: "file",
			args: []string{"file.txt"},
			notExists: []string{"file.txt"},
		},
		{
			name: "symlink",
			args: []string{"symlink"},
			notExists: []string{"symlink"},
		},
		{
			name: "error: dir",
			args: []string{"emptydir"},
			wantsErr: true,
		},
		{
			name: "empty dir with -r",
			args: []string{"-r", "emptydir"},
			notExists: []string{"emptydir"},
		},
		{
			name: "not empty dir with -r",
			args: []string{"-r", "notemptydir"},
			notExists: []string{"notemptydir"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tish.CreateTestFolders(t, "mkdir", map[string]string{
				"file.txt":             "hello world",
				"emptydir/":            "",
				"notemptydir/test.txt": "test",
				"@symlink":             "notemptydir",
			})

			s := tish.NewShell(root, []string{})
			md := rmCommand()

			p := tish.NewProcess(s, md.Executor, md.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())

			if tt.wantsErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, ne := range tt.notExists {
					assert.True(t, osutils.FileNotExists(s.ExpandPath(ne)))
				}
			}
		})
	}
}
