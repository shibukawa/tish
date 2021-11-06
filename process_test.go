package tish

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcess_RedirectStdin(t *testing.T) {
	root := CreateTestFolders(t, "redirect-stdout", map[string]string{"test.txt": "hello world"})

	s := NewShell(root, []string{})
	d := mockExecutor{}
	p := NewProcess(s, d.Executor, "mock", nil, 10, 11, nil)
	err := p.RedirectStdin(filepath.Join(root, "test.txt"))
	assert.NoError(t, err)
	err = p.Start(context.Background())
	assert.NoError(t, err)
	err = p.Wait()
	assert.NoError(t, err)
	assert.Equal(t, "hello world", d.Stdin)
}

func TestProcess_RedirectStdout(t *testing.T) {
	root := CreateTestFolders(t, "redirect-stdout", map[string]string{"test.txt": "hello world"})
	type args struct {
		append bool
		path   string
		output string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success: no append, new file",
			args: args{
				append: false,
				path:   filepath.Join(root, "new.txt"),
				output: "good night",
			},
			want:    "good night",
			wantErr: false,
		},
		{
			name: "success: append, existing file",
			args: args{
				append: true,
				path:   filepath.Join(root, "test.txt"),
				output: "\ngood night",
			},
			want:    "hello world\ngood night",
			wantErr: false,
		},
		{
			name: "success: append, new file",
			args: args{
				append: true,
				path:   filepath.Join(root, "new2.txt"),
				output: "good night",
			},
			want:    "good night",
			wantErr: false,
		},
		{
			name: "error: can't create file in non existing folder",
			args: args{
				append: false,
				path:   filepath.Join(root, "not-exist", "bad.txt"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShell(root, []string{})
			d := mockExecutor{
				Stdout: tt.args.output,
			}
			p := NewProcess(s, d.Executor, "mock", nil, 10, 11, nil)
			err := p.RedirectStdout(tt.args.path, tt.args.append)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				err = p.Start(context.Background())
				assert.NoError(t, err)
				err = p.Wait()
				assert.NoError(t, err)
				c, err := ioutil.ReadFile(tt.args.path)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(c))
			}
		})
	}
}

func TestProcess_RedirectStderr(t *testing.T) {
	root := CreateTestFolders(t, "redirect-stderr", map[string]string{"test.txt": "hello world"})

	type args struct {
		append bool
		path   string
		output string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success: no append, new file",
			args: args{
				append: false,
				path:   filepath.Join(root, "new.txt"),
				output: "good night",
			},
			want:    "good night",
			wantErr: false,
		},
		{
			name: "success: append, existing file",
			args: args{
				append: true,
				path:   filepath.Join(root, "test.txt"),
				output: "\ngood night",
			},
			want:    "hello world\ngood night",
			wantErr: false,
		},
		{
			name: "success: append, new file",
			args: args{
				append: true,
				path:   filepath.Join(root, "new2.txt"),
				output: "good night",
			},
			want:    "good night",
			wantErr: false,
		},
		{
			name: "error: can't create file in non existing folder",
			args: args{
				append: false,
				path:   filepath.Join(root, "not-exist", "bad.txt"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShell(root, []string{})
			d := mockExecutor{
				Stderr: tt.args.output,
			}
			p := NewProcess(s, d.Executor, "mock", nil, 10, 11, nil)
			err := p.RedirectStderr(tt.args.path, tt.args.append)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				err = p.Start(context.Background())
				assert.NoError(t, err)
				err = p.Wait()
				assert.NoError(t, err)
				c, err := ioutil.ReadFile(tt.args.path)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, string(c))
			}
		})
	}
}

func TestProcess_Pipe(t *testing.T) {
	s := NewShell(".", []string{})
	senderExecutor := mockExecutor{
		Stdout: "hello",
	}
	receiverExecutor := mockExecutor{}

	sender := NewProcess(s, senderExecutor.Executor, "mock", nil, 10, 11, nil)
	receiver := NewProcess(s, receiverExecutor.Executor, "child", nil, 10, 12, nil)

	sender.Pipe(receiver)

	err := sender.Start(context.Background())
	assert.NoError(t, err)

	err = receiver.Start(context.Background())
	assert.NoError(t, err)

	err = sender.Wait()
	assert.NoError(t, err)

	err = receiver.Wait()
	assert.NoError(t, err)

	assert.Equal(t, receiverExecutor.Stdin, senderExecutor.Stdout)
}
