package tish

import (
	"strings"
	"testing"
	"time"
)

func TestPrompt(t *testing.T) {
	type args struct {
		user   string
		host   string
		wd    string
		home  string
		now    time.Time
		status int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "contains time",
			args: args{
				now:  time.Date(2021, time.April, 10, 16, 17, 18, 00, time.Local),
			},
			want: "🕓16:17:18",
		},
		{
			name: "contains user and host",
			args: args{
				user: "myname",
				host: "host",
			},
			want: "myname@host",
		},
		{
			name: "contains abs dir if working directory is not under home",
			args: args{
				wd: "/tmp",
				home: "/users/home",
			},
			want: "📁/tmp ",
		},
		{
			name: "contains rel dir if working directory is under home",
			args: args{
				wd: "/users/home/sample",
				home: "/users/home",
			},
			want: "📁~/sample ",
		},
		{
			name: "contains rel dir if working directory is under home(2)",
			args: args{
				wd: "/users/home/sample/sub/subsub",
				home: "/users/home",
			},
			want: "📁~/sample/sub/subsub ",
		},
		{
			name: "contains tilde if working directory is home",
			args: args{
				wd: "/users/home",
				home: "/users/home",
			},
			want: "📁~ ",
		},
		{
			name: "contains status if status is not 0",
			args: args{
				status: 0,
			},
			want: " ✔︎ ",
		},
		{
			name: "contains status if status is not 0",
			args: args{
				status: 1,
			},
			want: " ✘ ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Prompt(tt.args.user, tt.args.host, tt.args.wd, tt.args.home, tt.args.now, tt.args.status, true); !strings.Contains(got, tt.want) {
				t.Errorf("Prompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
