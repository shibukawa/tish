package ls

import (
	"bytes"
	"context"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_lsCommand(t *testing.T) {
	root := tish.CreateTestFolders(t, "ls", map[string]string{
		"test.txt":       "hello world",
		"dir/first.txt":  "hello world",
		"dir/second.txt": "hello world",
		".secret.txt":    "secret",
		"dest1/":         "",
		"@symlink":       "dir",
	})
	tests := []struct {
		name  string
		args  []string
		wants []string
	}{
		{
			name:  "standard ls on root",
			args:  []string{},
			wants: []string{"dest1", "dir", "symlink", "test.txt"},
		},
		{
			name:  "all",
			args:  []string{"-a"},
			wants: []string{".", "..", ".secret.txt", "dest1", "dir", "symlink", "test.txt"},
		},
		{
			name:  "all without dot",
			args:  []string{"-A"},
			wants: []string{".secret.txt", "dest1", "dir", "symlink", "test.txt"},
		},
		{
			name:  "specific files",
			args:  []string{".secret.txt", "test.txt"},
			wants: []string{".secret.txt", "test.txt"},
		},
		{
			name:  "specify dir",
			args:  []string{"dir"},
			wants: []string{"first.txt", "second.txt"},
		},
		{
			name: "list",
			args: []string{"-l"},
			wants: []string{
				"drwxr-xr-x                               64 * dest1",
				"drwxr-xr-x                              128 * dir",
				"Lrwxr-xr-x                               84 * symlink",
				"-rwxr-xr-x                               11 * test.txt",
			},
		},
		{
			name:  "list: follow link",
			args:  []string{"-lL"},
			wants: []string{
				"drwxr-xr-x                               64 * dest1",
				"drwxr-xr-x                              128 * dir",
				"drwxr-xr-x                              128 * symlink", // same as dir
				"-rwxr-xr-x                               11 * test.txt",
			},
		},
		{
			name:  "specify symlink",
			args:  []string{"symlink"},
			wants: []string{"first.txt", "second.txt"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(root, []string{})
			l := lsCommand()

			var stdout bytes.Buffer
			p := tish.NewProcess(s, l.Executor, l.Name, tt.args, 10, 11, nil)
			p.Stdout = &stdout
			err := p.StartAndWait(context.Background())

			assert.NoError(t, err)

			var lines []string
			for _, entry := range strings.Split(stdout.String(), "\n") {
				if entry != "" {
					lines = append(lines, entry)
				}
			}
			assert.Equal(t, len(tt.wants), len(lines), "expect: %v, actual %v", tt.wants, lines)
			for i, l := range tt.wants {
				if len(tt.wants) == len(lines) {
					m, err := filepath.Match(l, lines[i])
					assert.NoError(t, err)
					assert.True(t, m, "should match: '%s' but '%s' doesn't", l, lines[i])
				}
			}
		})
	}
}
