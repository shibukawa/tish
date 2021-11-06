package tish

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func CreateTestFolders(t *testing.T, commandName string, files ...map[string]string) string {
	t.Helper()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	root := filepath.Join(os.TempDir(), fmt.Sprintf("tish-%s-test-%d", commandName, r.Int()))
	os.MkdirAll(root, 0755)
	if len(files) > 0 {
		for pathRule, content := range files[0] {
			var isSymLink bool
			if strings.HasPrefix(pathRule, "@") {
				isSymLink = true
				pathRule = strings.TrimPrefix(pathRule, "@")
			}
			f := strings.Split(pathRule, "|")
			path := f[0]
			var mode uint32
			if len(f) > 1 { //like |755
				m, err := strconv.ParseUint(f[1], 8, 32)
				if err != nil {
					t.Errorf("parse mode error: %s %w", pathRule, err)
					continue
				}
				mode = uint32(m)
			}
			dir := filepath.Dir(path)
			if dir != "." {
				if mode == 0 {
					mode = 0755
				}
				os.MkdirAll(filepath.Join(root, dir), 0755)
			}
			if isSymLink {
				os.Symlink(filepath.Join(root, content), filepath.Join(root, path))
			} else if !strings.HasSuffix(path, "/") {
				if mode == 0 {
					mode = 0777
				}
				ioutil.WriteFile(filepath.Join(root, path), []byte(content), os.FileMode(mode))
			}
		}
	}
	t.Cleanup(func() {
		os.RemoveAll(root)
	})
	return root
}

type mockExecutor struct {
	Stdin    string
	Stdout   string
	Stderr   string
	ExitCode int
	Callback func()
	Args     []string
}

func registerMockCommand(t *testing.T, s *Shell, name string) *mockExecutor {
	t.Helper()

	me := &mockExecutor{}

	dummyCommand := &Command{
		Name:      name,
		Executor:  me.Executor,
		Completer: nil,
	}
	s.commands = append(s.commands, dummyCommand)
	return me
}

func (d *mockExecutor) Executor(ctx context.Context, result *ExecResult, p *Process) (err error) {
	if d.Callback != nil {
		d.Callback()
	}
	d.Args = p.Args
	if p.Stdout != nil {
		_, err = io.WriteString(p.Stdout, d.Stdout)
		if err != nil {
			return err
		}
	}
	if p.Stderr != nil {
		_, err = io.WriteString(p.Stderr, d.Stderr)
		if err != nil {
			return err
		}
	}
	if p.Stdin != nil {
		c, err := ioutil.ReadAll(p.Stdin)
		if err != nil {
			return err
		}
		d.Stdin = string(c)
	}
	result.SetInternalProcessResult(d.ExitCode)
	return nil
}
