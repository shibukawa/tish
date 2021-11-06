package cp

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

func Test_cpCommand(t *testing.T) {
	type wants struct {
		exists    []string
		err       bool
	}
	tests := []struct {
		name  string
		args  []string
		wants wants
	}{
		{
			name: "copy single file to file",
			args: []string{"test.txt", "dest1/test.txt"},
			wants: wants{
				exists:    []string{"dest1/test.txt"},
			},
		},
		{
			name: "copy single file to dir",
			args: []string{"test.txt", "dest1/"},
			wants: wants{
				exists:    []string{"dest1/test.txt"},
			},
		},
		{
			name: "error: copy single file to no exist dir",
			args: []string{"test.txt", "dest2/test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "copy dir to dir with recursive flag",
			args: []string{"-R", "dir", "dest1"},
			wants: wants{
				exists:    []string{"dest1/dir/first.txt", "dest1/dir/second.txt"},
			},
		},
		{
			name: "error: copy dir to dir without recursive flag",
			args: []string{"dir", "dest1"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "error copy dir to file",
			args: []string{"dir", "test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "error copy multiple files to file",
			args: []string{"dir/first.txt", "dir/second.txt", "test.txt"},
			wants: wants{
				err: true,
			},
		},
		{
			name: "copy multiple files to dir",
			args: []string{"dir/first.txt", "dir/second.txt", "dest1"},
			wants: wants{
				exists:    []string{"dest1/first.txt", "dest1/second.txt"},
			},
		},
		{
			name: "copy multiple files to dir recursively",
			args: []string{"-R", "dir", "dest1"},
			wants: wants{
				exists:    []string{"dest1/dir/first.txt", "dest1/dir/second.txt"},
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
			c := CpCommand()

			p := tish.NewProcess(s, c.Executor, c.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())
			if tt.wants.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, p := range tt.wants.exists {
					assert.FileExists(t, filepath.Join(root, p))
				}
			}
		})
	}
}

