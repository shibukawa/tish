package sleep

import (
	"context"
	"testing"
	"time"

	"github.com/shibukawa/tish"

	"github.com/stretchr/testify/assert"
)

func Test_sleepCommand(t *testing.T) {
	t.Parallel()
	type args struct {
		args []string
	}
	type wants struct {
		duration time.Duration
	}
	tests := []struct {
		name  string
		args  args
		wants wants
	}{
		{
			name: "sleep",
			args: args{
				args: []string{"0.1"},
			},
			wants: wants{
				duration: 100 * time.Millisecond,
			},
		},
		{
			name: "sleep with non number duration",
			args: args{
				args: []string{"0.1s"},
			},
			wants: wants{
				duration: 100 * time.Millisecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tish.NewShell("/dummy", nil)
			sleep := SleepCommand()
			p := tish.NewProcess(s, sleep.Executor, sleep.Name, tt.args.args, 10, 11, nil)
			p.StartAndWait(context.Background())
			assert.True(t, p.Result.WallTime() < tt.wants.duration*12/10)
			assert.True(t, tt.wants.duration*8/10 < p.Result.WallTime())
		})
	}
}
