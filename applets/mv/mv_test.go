package mv

import (
	"context"
	"log"
	"path/filepath"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_mvCommand(t *testing.T) {
	type wants struct {
		exists    []string
		notExists []string
		err       bool
	}
	tests := []struct {
		name  string
		args  []string
		wants wants
	}{
		{
			name: "move single file to file",
			args: []string{"test.txt", "dest1/test.txt"},
			wants: wants{
				exists:    []string{"dest1/test.txt"},
				notExists: []string{"test.txt"},
			},
		},
		{
			name: "move single file to dir (1)",
			args: []string{"test.txt", "dest1"},
			wants: wants{
				exists:    []string{"dest1/test.txt"},
				notExists: []string{"test.txt"},
			},
		},
		{
			name: "move single file to dir (2)",
			args: []string{"test.txt", "dest1/"},
			wants: wants{
				exists:    []string{"dest1/test.txt"},
				notExists: []string{"test.txt"},
			},
		},
		{
			name: "error: move single file to no exist dir",
			args: []string{"test.txt", "dest2/test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "move dir to dir",
			args: []string{"dir", "dest1"},
			wants: wants{
				exists:    []string{"dest1/dir/first.txt", "dest1/dir/second.txt"},
				notExists: []string{"dir/first.txt", "dir/second.txt"},
			},
		},
		{
			name: "move dir to new dir",
			args: []string{"dir", "dest2"},
			wants: wants{
				exists:    []string{"dest2/first.txt", "dest2/second.txt"},
				notExists: []string{"dir/first.txt", "dir/second.txt"},
			},
		},
		{
			name: "error move dir to file",
			args: []string{"dir", "test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "error move multiple files to file",
			args: []string{"dir/first.txt", "dir/second.txt", "test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "move multiple files to dir",
			args: []string{"dir/first.txt", "dir/second.txt", "dest1"},
			wants: wants{
				exists:    []string{"dest1/first.txt", "dest1/second.txt"},
				notExists: []string{"dir/first.txt", "dir/second.txt"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tish.CreateTestFolders(t, "mv", map[string]string{
				"test.txt":       "hello world",
				"dir/first.txt":  "hello world",
				"dir/second.txt": "hello world",
				"dest1/":         "",
			})

			s := tish.NewShell(root, []string{})
			m := MvCommand()

			p := tish.NewProcess(s, m.Executor, m.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())
			if tt.wants.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, p := range tt.wants.exists {
					assert.FileExists(t, filepath.Join(root, p))
				}
				for _, p := range tt.wants.notExists {
					assert.NoFileExists(t, filepath.Join(root, p))
				}
			}
		})
	}
}
