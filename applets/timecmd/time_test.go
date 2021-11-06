package timecmd

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/shibukawa/tish"
	"github.com/stretchr/testify/assert"

	_ "github.com/shibukawa/tish/applets/sleep"
)

func Test_timeCommand(t *testing.T) {
	t.Parallel()
	type args struct {
		args []string
	}
	type wants struct {
		pattern *regexp.Regexp
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "time",
			args: args{
				args: []string{"sleep", "0.1"},
			},
			wants: wants{
				pattern: regexp.MustCompile(`sleep 0.1  0.00s user 0.00s system 0% cpu 0.1\d* total`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/dummy", []string{})
			timecmd := TimeCommand()
			var stdout bytes.Buffer
			p := tish.NewProcess(s, timecmd.Executor, timecmd.Name, tt.args.args, 10, 11, nil)
			p.Stdout = &stdout
			err := p.StartAndWait(context.Background())
			assert.NoError(t, err)
			assert.True(t, tt.wants.pattern.MatchString(stdout.String()))
		})
	}
}
