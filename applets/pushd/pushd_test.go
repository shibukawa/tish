package pushd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"
)

func Test_pushdAndPopdCommand(t *testing.T) {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("tish-pushd-popd-test-%d", rand.Int()))
	os.MkdirAll(filepath.Join(root, "current", "sub1", "sub2"), 0o777)
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	type args struct {
		pushd []string
		popd  int
	}
	type wants struct {
		wd          string
		pushdStatus int
		popdStatus  int
		stdout      string
		stderr      string
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "relative/pushd",
			args: args{
				pushd: []string{".."},
				popd:  0,
			},
			wants: wants{
				wd:          root,
				pushdStatus: 0,
				popdStatus:  0,
				stdout:      root + " ~\n",
				stderr:      "",
			},
		},
		{
			name: "relative/pushd multiple",
			args: args{
				pushd: []string{"sub1", "sub2"},
				popd:  0,
			},
			wants: wants{
				wd:          filepath.Join(root, "current", "sub1", "sub2"),
				pushdStatus: 0,
				popdStatus:  0,
				stdout:      "~/sub1 ~\n~/sub1/sub2 ~/sub1 ~\n",
				stderr:      "",
			},
		},
		{
			name: "absolute/pushd",
			args: args{
				pushd: []string{root},
				popd:  0,
			},
			wants: wants{
				wd:          root,
				pushdStatus: 0,
				popdStatus:  0,
				stdout:      root + " ~\n",
				stderr:      "",
			},
		},
		{
			name: "relative/pushd and popd",
			args: args{
				pushd: []string{"sub1"},
				popd:  1,
			},
			wants: wants{
				wd:          filepath.Join(root, "current"),
				pushdStatus: 0,
				popdStatus:  0,
				stdout:      "~/sub1 ~\n~\n",
				stderr:      "",
			},
		},
		{
			name: "relative/pushd and popd multiple",
			args: args{
				pushd: []string{"sub1", "sub2"},
				popd:  2,
			},
			wants: wants{
				wd:          filepath.Join(root, "current"),
				pushdStatus: 0,
				popdStatus:  0,
				stdout:      "~/sub1 ~\n~/sub1/sub2 ~/sub1 ~\n~/sub1 ~\n~\n",
				stderr:      "",
			},
		},
		{
			name: "relative/pushd error",
			args: args{
				pushd: []string{"not_found"},
				popd:  0,
			},
			wants: wants{
				wd:          filepath.Join(root, "current"),
				pushdStatus: 1,
				popdStatus:  0,
				stdout:      "",
				stderr:      "pushd: no such file or directory: not_found\n",
			},
		},
		{
			name: "relative/popd error (too many pop)",
			args: args{
				pushd: []string{"sub1"},
				popd:  2,
			},
			wants: wants{
				wd:          filepath.Join(root, "current"),
				pushdStatus: 0,
				popdStatus:  1,
				stdout:      "~/sub1 ~\n~\n",
				stderr:      "popd: directory stack empty\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(filepath.Join(root, "current"), []string{
				"HOME=" + filepath.Join(root, "current"),
				"USERPROFILE=" + filepath.Join(root, "current"),
			})
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			t.Run("pushd", func(t *testing.T) {
				var status int
				pushd := pushdCommand()
				for _, dir := range tt.args.pushd {
					p1 := tish.NewProcess(s, pushd.Executor, pushd.Name, []string{dir}, 10, 11, nil)
					p1.Stdout = &stdout
					p1.Stderr = &stderr
					p1.StartAndWait(context.Background())
					status = p1.Result.ExitCode()
				}
				assert.Equal(t, tt.wants.pushdStatus, status)
			})
			t.Run("popd", func(t *testing.T) {
				var status int
				popd := popdCommand()
				for i := 0; i < tt.args.popd; i++ {
					p1 := tish.NewProcess(s, popd.Executor, popd.Name, []string{}, 10, 11, nil)
					p1.Stdout = &stdout
					p1.Stderr = &stderr
					p1.StartAndWait(context.Background())
					status = p1.Result.ExitCode()
				}
				assert.Equal(t, tt.wants.popdStatus, status)
			})
			assert.Equal(t, tt.wants.wd, s.WorkingDir())
			assert.Equal(t, tt.wants.stdout, stdout.String())
			assert.Equal(t, tt.wants.stderr, stderr.String())
		})
	}
}

func Test_dirsCommand(t *testing.T) {
	root := filepath.Join(os.TempDir(), fmt.Sprintf("tish-dirs-test-%d", rand.Int()))
	os.MkdirAll(filepath.Join(root, "current", "sub1", "sub2"), 0o777)
	os.MkdirAll(filepath.Join(root, "other", "sub1", "sub2"), 0o777)
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	envs := []string{
		"HOME=" + filepath.Join(root, "current"),
		"USERPROFILE=" + filepath.Join(root, "current"),
	}
	type args struct {
		wd    string
		pushd []string
		popd  int
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
			name: "empty at home",
			args: args{
				pushd: []string{},
			},
			wants: wants{
				stdout: "~\n",
			},
		},
		{
			name: "subdir under home",
			args: args{
				pushd: []string{"sub1"},
			},
			wants: wants{
				stdout: "~/sub1 ~\n",
			},
		},
		{
			name: "subdirs under home",
			args: args{
				pushd: []string{"sub1", "sub2"},
			},
			wants: wants{
				stdout: "~/sub1/sub2 ~/sub1 ~\n",
			},
		},
		{
			name: "dir outside of home",
			args: args{
				pushd: []string{filepath.Join(root, "other")},
			},
			wants: wants{
				stdout: filepath.Join(root, "other") + " ~\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell(filepath.Join(root, "current"), envs)
			var status int
			pushd := pushdCommand()
			for _, dir := range tt.args.pushd {
				p1 := tish.NewProcess(s, pushd.Executor, pushd.Name, []string{dir}, 10, 11, nil)
				p1.Stdout = io.Discard
				p1.Stderr = io.Discard
				p1.StartAndWait(context.Background())
				status = p1.Result.ExitCode()
			}
			assert.Equal(t, 0, status)
			dirs := dirsCommand()
			var stdout bytes.Buffer
			p2 := tish.NewProcess(s, dirs.Executor, dirs.Name, nil, 10, 11, nil)
			p2.Stdout = &stdout
			p2.StartAndWait(context.Background())
			assert.Equal(t, 0, p2.Result.ExitCode())
			assert.Equal(t, tt.wants.stdout, stdout.String())
		})
	}
}
