package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"os/user"
	"time"

	"github.com/fatih/color"
	"github.com/shibukawa/tish"
	_ "github.com/shibukawa/tish/applets"
	"github.com/peterh/liner"
)

var (
	dirColor = color.New(color.FgHiGreen).SprintFunc()
	cursorColor = color.New(color.FgCyan, color.Bold).SprintFunc()
)

func init() {
	log.SetPrefix("🐙 ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func completor(input string) []string {
	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get working directory: %v\n", err)
		os.Exit(1)
	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get home directory: %v\n", err)
		os.Exit(1)
	}
	user, err := user.Current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get user name: %v\n", err)
		os.Exit(1)
	}
	hostName, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get host name: %v\n", err)
		os.Exit(1)
	}

	shell := tish.NewShell(wd, os.Environ())
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)
	line.SetCompleter(completor)
	defer line.Close()
	// pushd := []string{wd}

	color.New(color.FgYellow).Println("🐸 tiny shell")
	lastStatus := 0
	for {
		wd = shell.WorkingDir()

		fmt.Printf("\n" + tish.Prompt(user.Username, hostName, wd, homedir, time.Now(), lastStatus, false))
		if cmd, err := line.Prompt(" "); err == nil {
			if cmd == "" {
				continue
			}
			status, err := shell.Run(ctx, cmd, os.Stdout, os.Stderr)
			if errors.Is(err, tish.ErrExit) {
				os.Exit(status)
				break
			} else if err != nil {
				log.Print("Error reading line: ", err)
			}
			line.AppendHistory(cmd)
		} else if errors.Is(err, io.EOF) {
			break
		} else if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			break
		} else {
			log.Print("Error reading line: ", err)
		}
	}
}
