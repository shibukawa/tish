package parser

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCommandStr(t *testing.T) {
	type args struct {
		cmdStr string
	}
	tests := []struct {
		name string
		args args
		want [][]*Session
	}{
		{
			name: "single command",
			args: args{
				cmdStr: `time`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "time",
							},
						},
					},
				},
			},
		},
		{
			name: "single command with args",
			args: args{
				cmdStr: `sleep 10`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "sleep",
							},
							{
								Term: "10",
							},
						},
					},
				},
			},
		},
		{
			name: "single command with quoted args",
			args: args{
				cmdStr: `sleep "10"`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "sleep",
							},
							{
								Term: "10",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommandStr(tt.args.cmdStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCommandStr_Separator(t *testing.T) {
	type args struct {
		cmdStr string
	}
	tests := []struct {
		name string
		args args
		want [][]*Session
	}{
		{
			name: "commands with pipe",
			args: args{
				cmdStr: `cat sample.txt | wc`,
			},
			want: [][]*Session{ // [ [proc1, proc2] ]
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
					},
					{
						Fragments: []Fragment{
							{
								Term: "wc",
							},
						},
					},
				},
			},
		},
		{
			name: "commands with semicolon", // [ [proc1], [proc2] ]
			args: args{
				cmdStr: `cat sample.txt ; date`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
					},
				},
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
					},
				},
			},
		},
		{
			name: "commands with logical or", // [ [proc1], [proc2] ]
			args: args{
				cmdStr: `cat sample.txt || date`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
						Separator: LogicalOr,
					},
				},
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
					},
				},
			},
		},
		{
			name: "commands with logical and", // [ [proc1], [proc2] ]
			args: args{
				cmdStr: `cat sample.txt && date`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
						Separator: LogicalAnd,
					},
				},
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommandStr(tt.args.cmdStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCommandStr_Redirect(t *testing.T) {
	type args struct {
		cmdStr string
	}
	tests := []struct {
		name string
		args args
		want [][]*Session
	}{
		{
			name: "single command",
			args: args{
				cmdStr: `time`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "time",
							},
						},
					},
				},
			},
		},
		{
			name: "single command with args",
			args: args{
				cmdStr: `sleep 10`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "sleep",
							},
							{
								Term: "10",
							},
						},
					},
				},
			},
		},
		{
			name: "single command with quoted args",
			args: args{
				cmdStr: `sleep "10"`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "sleep",
							},
							{
								Term: "10",
							},
						},
					},
				},
			},
		},
		{
			name: "commands with pipe",
			args: args{
				cmdStr: `cat sample.txt | wc`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
					},
					{
						Fragments: []Fragment{
							{
								Term: "wc",
							},
						},
					},
				},
			},
		},
		{
			name: "commands with semicolon",
			args: args{
				cmdStr: `cat sample.txt ; date`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "cat",
							},
							{
								Term: "sample.txt",
							},
						},
					},
				},
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
					},
				},
			},
		},
		{
			name: "redirect: stdin",
			args: args{
				cmdStr: `wc < file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "wc",
							},
						},
						Stdin: "file.txt",
					},
				},
			},
		},
		{
			name: "redirect: stdout",
			args: args{
				cmdStr: `date > file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stdout: "file.txt",
					},
				},
			},
		},
		{
			name: "redirect: stdout(append)",
			args: args{
				cmdStr: `date >> file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stdout:       "file.txt",
						StdoutAppend: true,
					},
				},
			},
		},
		{
			name: "redirect: stderr",
			args: args{
				cmdStr: `date 2> file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stderr: "file.txt",
					},
				},
			},
		},
		{
			name: "redirect: stderr (append)",
			args: args{
				cmdStr: `date 2>> file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stderr:       "file.txt",
						StderrAppend: true,
					},
				},
			},
		},
		{
			name: "redirect: stdout and stderr",
			args: args{
				cmdStr: `date &> file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stdout: "file.txt",
						Stderr: "file.txt",
					},
				},
			},
		},
		{
			name: "redirect: stdout and stderr",
			args: args{
				cmdStr: `date &>> file.txt`,
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "date",
							},
						},
						Stdout:       "file.txt",
						StdoutAppend: true,
						Stderr:       "file.txt",
						StderrAppend: true,
					},
				},
			},
		},
		{
			name: "back quote: simple",
			args: args{
				cmdStr: "echo `date`",
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "echo",
							},
							{
								Term: "",
								Sessions: []*Session{
									{
										Fragments: []Fragment{
											{
												Term: "date",
											},
										},
									},
								},
								Texts: []string{""},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommandStr(tt.args.cmdStr)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCommandStr_Error(t *testing.T) {
	type args struct {
		cmdStr string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]*Session
		wantErr bool
	}{
		{
			name: "redirect pipe",
			args: args{
				cmdStr: "echo > |",
			},
		},
		{
			name: "no redirect target",
			args: args{
				cmdStr: "echo >",
			},
		},
		{
			name: "no command after pipe",
			args: args{
				cmdStr: "echo |",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCommandStr(tt.args.cmdStr)
			assert.Error(t, err)
		})
	}
}

func TestParseCommandStr_Backquote(t *testing.T) {
	type args struct {
		cmdStr string
	}
	tests := []struct {
		name    string
		args    args
		want    [][]*Session
		wantErr bool
	}{
		{
			name: "single backquote",
			args: args{
				cmdStr: "echo `date`",
			},
			want: [][]*Session{
				{
					{
						Fragments: []Fragment{
							{
								Term: "echo",
							},
							{
								Texts: []string{""},
								Sessions: []*Session{
									{
										Fragments: []Fragment{
											{
												Term: "date",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommandStr(tt.args.cmdStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommandStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseBackquote(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no backquote",
			args: args{
				source: "no back quote string",
			},
			want: []string{
				"no back quote string",
			},
		},
		{
			name: "backquote",
			args: args{
				source: "`back quote string`",
			},
			want: []string{
				"", "back quote string",
			},
		},
		{
			name: "no close backquote",
			args: args{
				source: "`back quote is not closed",
			},
			wantErr: true,
		},
		{
			name: "several back quotes",
			args: args{
				source: "today is `date` current branch is `git branch`",
			},
			want: []string{
				"today is ", "date", " current branch is ", "git branch",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBackquote(tt.args.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseBackquote() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
