package chmod

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Test_chmodCommand(t *testing.T) {
	type wants struct {
		modes map[string]os.FileMode
		err   bool
	}
	tests := []struct {
		name  string
		args  []string
		wants wants
	}{
		{
			name: "single : +r",
			args: []string{"+r", "secret.txt"},
			wants: wants{
				modes: map[string]os.FileMode{
					"secret.txt": os.FileMode(0444),
				},
			},
		},
		{
			name: "single: +w",
			args: []string{"+w", "secret.txt"},
			wants: wants{
				modes: map[string]os.FileMode{
					"secret.txt": os.FileMode(0622),
				},
			},
		},
		{
			name: "single: +rx",
			args: []string{"+rx", "secret.txt"},
			wants: wants{
				modes: map[string]os.FileMode{
					"secret.txt": os.FileMode(0555),
				},
			},
		},
		{
			name: "recursive: +w",
			args: []string{"-R", "+w", "dir"},
			wants: wants{
				modes: map[string]os.FileMode{
					"dir/file.txt": os.FileMode(0666),
					"dir/exec":     os.FileMode(0777),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tish.CreateTestFolders(t, "mv", map[string]string{
				"secret.txt|400":   "hello world",
				"dir/file.txt|644": "hello world",
				"dir/exec|755":     "#!/bin/sh\necho 'hello'",
			})

			s := tish.NewShell(root, []string{})
			c := chmodCommand()

			p := tish.NewProcess(s, c.Executor, c.Name, tt.args, 10, 11, nil)
			err := p.StartAndWait(context.Background())
			if tt.wants.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for p, m := range tt.wants.modes {
					fi, err := os.Stat(filepath.Join(root, p))
					assert.NoError(t, err)
					assert.Equal(t, strconv.FormatUint(uint64(m), 8), strconv.FormatUint(uint64(fi.Mode()), 8))
				}
			}
		})
	}
}

func Test_parseFlag(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name     string
		args     string
		wantPlus uint32
		wantMask uint32
		wantErr  bool
	}{
		{
			name:     "octet characters: 1",
			args:     "777",
			wantPlus: 0777,
			wantMask: 0000,
		},
		{
			name:     "octet characters: 2",
			args:     "644",
			wantPlus: 0644,
			wantMask: 0000,
		},
		{
			name:     "octet characters: 3",
			args:     "666",
			wantPlus: 0666,
			wantMask: 0000,
		},
		{
			name:     "user +",
			args:     "u+x",
			wantPlus: 0100,
			wantMask: 0677,
		},
		{
			name:     "group +",
			args:     "g+x",
			wantPlus: 0010,
			wantMask: 0767,
		},
		{
			name:     "other +",
			args:     "o+x",
			wantPlus: 0001,
			wantMask: 0776,
		},
		{
			name:     "user and group +",
			args:     "ug+x",
			wantPlus: 0110,
			wantMask: 0667,
		},
		{
			name:     "user -",
			args:     "u-w",
			wantPlus: 0000,
			wantMask: 0577,
		},
		{
			name:     "group -",
			args:     "g-w",
			wantPlus: 0000,
			wantMask: 0757,
		},
		{
			name:     "other -",
			args:     "o-w",
			wantPlus: 0000,
			wantMask: 0775,
		},
		{
			name:     "group and other -",
			args:     "go-x",
			wantPlus: 0000,
			wantMask: 0766,
		},
		{
			name:     "user =",
			args:     "u=r",
			wantPlus: 0400,
			wantMask: 0077,
		},
		{
			name:     "group =",
			args:     "g=r",
			wantPlus: 0040,
			wantMask: 0707,
		},
		{
			name:     "other =",
			args:     "o=r",
			wantPlus: 0004,
			wantMask: 0770,
		},
		{
			name:     "user and other =",
			args:     "uo=r",
			wantPlus: 0404,
			wantMask: 0070,
		},
		{
			name:     "all +",
			args:     "+r",
			wantPlus: 0444,
			wantMask: 0333,
		},
		{
			name:     "all -",
			args:     "-w",
			wantPlus: 0000,
			wantMask: 0555,
		},
		{
			name:     "all =",
			args:     "=x",
			wantPlus: 0111,
			wantMask: 0000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPlus, gotMinus, err := parseFlag(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPlus != tt.wantPlus {
				t.Errorf("parseFlag() gotPlus = %v, want %v", gotPlus, tt.wantPlus)
			}
			if gotMinus != tt.wantMask {
				t.Errorf("parseFlag() gotMinus = %v, want %v", gotMinus, tt.wantMask)
			}
		})
	}
}
