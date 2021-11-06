package wc

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/shibukawa/tish"
	"github.com/jessevdk/go-flags"
)

func init() {
	tish.RegisterCommand(WordCountCommand())
}

type config struct {
	Char bool `short:"c" description:"Write to the standard output the number of bytes in each input file."`
	Line bool `short:"l" description:"Write to the standard output the number of <newline> characters in each input file."`
	Word bool `short:"w" description:"Write to the standard output the number of words in each input file."`
}

func WordCountCommand() *tish.Command {
	return &tish.Command{
		Name: "wc",
		Executor: func(ctx context.Context, res *tish.ExecResult, p *tish.Process) (err error) {
			c := &config{}

			args, err := flags.ParseArgs(c, p.Args)
			if err != nil {
				res.SetInternalProcessResult(1)
				return nil
			}
			if len(args) == 0 || args[0] == "-" {
				if err := wc(p.Stdout, "", p.Stdin, c); err != nil {
					res.SetInternalProcessResult(1)
					return err
				}
				res.SetInternalProcessResult(0)
				return nil
			}
			for _, path := range args {
				f, err := os.Open(p.Shell.ExpandPath(path))
				if err != nil {
					return err
				}
				defer f.Close()

				if err := wc(p.Stdout, path, f, c); err != nil {
					res.SetInternalProcessResult(1)
					return err
				}
			}
			res.SetInternalProcessResult(0)
			return nil
		},
		Completer: nil,
	}
}

type result struct {
	lines int
	words int
	bytes int
}

func wc(w io.Writer, path string, f io.Reader, c *config) error {
	var ret result
	reader := bufio.NewReaderSize(f, 4096)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		ret.lines++
		ret.bytes += len(line) + 1 // +1 means NewLine
		ret.words += len(bytes.Fields(line))
	}
	var out []string
	if c.Char {
		out = append(out, strconv.Itoa(ret.bytes))
	}
	if c.Line {
		out = append(out, strconv.Itoa(ret.lines))
	}
	if c.Word {
		out = append(out, strconv.Itoa(ret.words))
	}
	if len(out) == 0 { // no flag set
		out = append(out, strconv.Itoa(ret.lines))
		out = append(out, strconv.Itoa(ret.words))
		out = append(out, strconv.Itoa(ret.bytes))
	}
	if path != "" {
		out = append(out, path)
	}

	fmt.Fprintln(w, strings.Join(out, " "))

	return nil
}
