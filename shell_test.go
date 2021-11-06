package tish

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shibukawa/tish/parser"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func TestExecExternalCommand(t *testing.T) {
	var opt string
	var expectedExitCode int
	if runtime.GOOS == "windows" {
		opt = "/?"
		expectedExitCode = 0
	} else {
		opt = "--help"
		expectedExitCode = 64
	}

	s := NewShell(".", nil)
	// s.commands = append(s.commands, PassCommand(t))
	var stdout bytes.Buffer
	state, err := s.Run(context.Background(), "ping " + opt, &stdout, &stdout)
	assert.Equal(t, expectedExitCode, state)
	assert.Error(t, err)
	assert.NotEmpty(t, stdout.String())
}

func TestShell_runSessionGroup_Pipe(t *testing.T) {
	root := CreateTestFolders(t, "run-session-group-pipe")

	s := NewShell(root, []string{})
	mock1 := registerMockCommand(t, s, "mock1")
	mock1.Stdout = "hello"
	mock2 := registerMockCommand(t, s, "mock2")

	sessions := []*parser.Session{
		{
			Fragments: []parser.Fragment{
				{
					Term: "mock1",
				},
			},
		},
		{
			Fragments: []parser.Fragment{
				{
					Term: "mock2",
				},
			},
		},
	}
	result, err := s.runSessionGroup(context.Background(), sessions, io.Discard, io.Discard)

	assert.NotNil(t, result)
	assert.NoError(t, err)
	assert.Equal(t, "hello", mock2.Stdin)
}

func TestShell_runSessionGroup_RedirectStdout(t *testing.T) {
	root := CreateTestFolders(t, "run-session-group", map[string]string{
		"test1.txt": "hello world\n",
		"test2.txt": "hello world\n",
	})

	s := NewShell(root, []string{})
	mock1 := registerMockCommand(t, s, "mock1")
	mock1.Stdout = "good night"

	sessions := []*parser.Session{
		{
			Fragments: []parser.Fragment{
				{
					Term: "mock1",
				},
			},
		},
	}

	type args struct {
		filename string
		append   bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no append",
			args: args{
				filename: "test1.txt",
				append:   false,
			},
			want: "good night",
		},
		{
			name: "append",
			args: args{
				filename: "test2.txt",
				append:   true,
			},
			want: "hello world\ngood night",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessions[0].Stdout = tc.args.filename
			sessions[0].StdoutAppend = tc.args.append
			result, err := s.runSessionGroup(context.Background(), sessions, io.Discard, io.Discard)

			assert.NotNil(t, result)
			assert.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
			content, _ := ioutil.ReadFile(filepath.Join(root, tc.args.filename))
			assert.Equal(t, tc.want, string(content))
		})
	}
}

func TestShell_runSessionGroup_RedirectStderr(t *testing.T) {
	root := CreateTestFolders(t, "run-session-group", map[string]string{
		"test1.txt": "hello world\n",
		"test2.txt": "hello world\n",
	})

	s := NewShell(root, []string{})
	mock1 := registerMockCommand(t, s, "mock1")
	mock1.Stderr = "good night"

	sessions := []*parser.Session{
		{
			Fragments: []parser.Fragment{
				{
					Term: "mock1",
				},
			},
		},
	}

	type args struct {
		filename string
		append   bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no append",
			args: args{
				filename: "test1.txt",
				append:   false,
			},
			want: "good night",
		},
		{
			name: "append",
			args: args{
				filename: "test2.txt",
				append:   true,
			},
			want: "hello world\ngood night",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessions[0].Stderr = tc.args.filename
			sessions[0].StderrAppend = tc.args.append
			result, err := s.runSessionGroup(context.Background(), sessions, io.Discard, io.Discard)

			assert.NotNil(t, result)
			assert.NoError(t, err)
			time.Sleep(100 * time.Millisecond)
			content, _ := ioutil.ReadFile(filepath.Join(root, tc.args.filename))
			assert.Equal(t, tc.want, string(content))
		})
	}
}

func TestShell_runSessionGroup_RedirectStdin(t *testing.T) {
	root := CreateTestFolders(t, "run-session-group", map[string]string{
		"test1.txt": "hello world\n",
	})

	s := NewShell(root, []string{})
	mock1 := registerMockCommand(t, s, "mock1")

	sessions := []*parser.Session{
		{
			Fragments: []parser.Fragment{
				{
					Term: "mock1",
				},
			},
		},
	}

	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no append",
			args: args{
				filename: "test1.txt",
			},
			want: "hello world\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sessions[0].Stdin = tc.args.filename
			result, err := s.runSessionGroup(context.Background(), sessions, io.Discard, io.Discard)

			assert.NotNil(t, result)
			assert.NoError(t, err)
			// time.Sleep(100 * time.Millisecond)
			assert.Equal(t, tc.want, mock1.Stdin)
		})
	}
}

func TestShell_Wildcard(t *testing.T) {
	root := CreateTestFolders(t, "wildcard", map[string]string{
		"test11.txt": "",
		"test12.txt": "",
		"test21.txt": "",
		"test22.txt": "",
	})

	type args struct {
		pattern string
	}

	tests := []struct {
		name      string
		args      args
		wantCount int
		wantErr   bool
	}{
		{
			name: "no wildcard",
			args: args{
				pattern: "test11.txt",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "question wildcard",
			args: args{
				pattern: "test1?.txt",
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "asterisk wildcard",
			args: args{
				pattern: "test*.txt",
			},
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "no match error",
			args: args{
				pattern: "test*.go",
			},
			wantCount: 0,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewShell(root, []string{})
			f := parser.Fragment{
				Term: tt.args.pattern,
			}
			files, err := s.expandWildcard(f)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, len(files))
			}
		})
	}
}

func TestShell_runSessionGroup_Separator(t *testing.T) {
	root := CreateTestFolders(t, "run-session-group")

	s := NewShell(root, []string{})
	mock1 := registerMockCommand(t, s, "mock1")
	mock2 := registerMockCommand(t, s, "mock2")

	sessions := [][]*parser.Session{
		{
			{
				Fragments: []parser.Fragment{
					{
						Term: "mock1",
					},
				},
			},
		},
		{
			{
				Fragments: []parser.Fragment{
					{
						Term: "mock2",
					},
				},
			},
		},
	}

	type args struct {
		sep parser.Separator
		mock1exitCode int
	}
	tests := []struct {
		name string
		args args
		wantCalled bool
	}{
		{
			name: "semicolon: first command success: called",
			args: args{
				sep: parser.Semicolon,
				mock1exitCode: 0,
			},
			wantCalled: true,
		},
		{
			name: "semicolon: first command fail: called",
			args: args{
				sep: parser.Semicolon,
				mock1exitCode: 1,
			},
			wantCalled: true,
		},
		{
			name: "logical or: first command success: not called",
			args: args{
				sep: parser.LogicalOr,
				mock1exitCode: 0,
			},
			wantCalled: false,
		},
		{
			name: "logical or: first command fail: called",
			args: args{
				sep: parser.LogicalOr,
				mock1exitCode: 1,
			},
			wantCalled: true,
		},
		{
			name: "logical and: first command success: called",
			args: args{
				sep: parser.LogicalAnd,
				mock1exitCode: 0,
			},
			wantCalled: true,
		},
		{
			name: "logical and: first command fail: not called",
			args: args{
				sep: parser.LogicalAnd,
				mock1exitCode: 1,
			},
			wantCalled: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock1.ExitCode = tc.args.mock1exitCode
			sessions[0][0].Separator = tc.args.sep
			var mock2Called bool
			mock2.Callback = func() {
				mock2Called = true
			}
			_, err := s.runSessionGroups(context.Background(), sessions, io.Discard, io.Discard)

			assert.NoError(t, err)
			assert.Equal(t, tc.wantCalled, mock2Called)
		})
	}
}
