package parser

import (
	"errors"
	"fmt"
	"github.com/google/shlex"
	"strings"
)

var (
	ErrBackquoteNotClosed         = errors.New("backquote not closed")
	ErrTooManySessionsInBackquote = errors.New("too many sessions in backquote")
	ErrNoRedirectTarget           = errors.New("no redirect target")
	ErrNoProcessAfterPipe         = errors.New("no process after pipe")
)

type Separator int

const (
	Semicolon  Separator = 0 // ;
	LogicalOr  Separator = 1 // ||
	LogicalAnd Separator = 2 // &&
)

type Session struct {
	Fragments    []Fragment
	Stdin        string
	Stdout       string
	StdoutAppend bool
	Stderr       string
	StderrAppend bool
	Separator    Separator
}

func (s Session) GetCommand() (cmd string, args []string) {
	cmd = s.Fragments[0].Term
	for _, f := range s.Fragments[1:] {
		args = append(args, f.Term)
	}
	return
}

func parseBackquote(source string) ([]string, error) {
	pos := 0
	start := 0
	inBackquote := false
	var result []string
	for pos != len(source) {
		if source[pos] == '`' {
			result = append(result, source[start:pos])
			start = pos + 1
			inBackquote = !inBackquote
		}
		pos++
	}
	if inBackquote {
		return nil, ErrBackquoteNotClosed
	}
	if pos != start {
		result = append(result, source[start:pos])
	}

	return result, nil
}

// Fragment represents small piece of word of commandline string.
//
// If the word has back quote, it returns sessions, texts.
//
//   "current branch is `git branch` and current commit is `git rev-parse HEAD`."
//
//   []Tests{"current branch is ", " and current commit is ", "."}
//   []Sessions{{"git", "branch"}, {"git", "rev-parse", "HEAD"}}
type Fragment struct {
	Term     string
	Sessions []*Session
	Texts    []string
}

func ParseCommandStr(cmdStr string) ([][]*Session, error) {
	lastSession := &Session{}
	pipedSessions := []*Session{lastSession}
	result := [][]*Session{pipedSessions}

	l := shlex.NewLexer(strings.NewReader(cmdStr))
	var nextStdin, nextStdout, nextStderr, nextStdoutAndStderr, nextAppend bool
	var lastField string
	for {
		var err error
		field, err := l.Next()
		if err != nil {
			break
		}
		lastField = field
		switch field {
		case "|":
			if nextStdin || nextStdout || nextStderr || nextStdoutAndStderr {
				return nil, fmt.Errorf("'%s' after redirect", field)
			}
			lastSession = &Session{}
			pipedSessions = append(pipedSessions, lastSession)
			result[len(result)-1] = pipedSessions
			continue
		case ";":
			if nextStdin || nextStdout || nextStderr || nextStdoutAndStderr {
				return nil, fmt.Errorf("'%s' after redirect", field)
			}
			lastSession = &Session{}
			pipedSessions = []*Session{lastSession}
			result = append(result, pipedSessions)
			continue
		case "||":
			if nextStdin || nextStdout || nextStderr || nextStdoutAndStderr {
				return nil, fmt.Errorf("'%s' after redirect", field)
			}
			lastSession.Separator = LogicalOr
			lastSession = &Session{}
			pipedSessions = []*Session{lastSession}
			result = append(result, pipedSessions)
			continue
		case "&&":
			if nextStdin || nextStdout || nextStderr || nextStdoutAndStderr {
				return nil, fmt.Errorf("'%s' after redirect", field)
			}
			lastSession.Separator = LogicalAnd
			lastSession = &Session{}
			pipedSessions = []*Session{lastSession}
			result = append(result, pipedSessions)
			continue
		case "<":
			nextStdin = true
			continue
		case ">":
			nextStdout = true
			nextAppend = false
			continue
		case ">>":
			nextStdout = true
			nextAppend = true
			continue
		case "2>":
			nextStderr = true
			nextAppend = false
			continue
		case "2>>":
			nextStderr = true
			nextAppend = true
			continue
		case "&>":
			nextStdoutAndStderr = true
			nextAppend = false
			continue
		case "&>>":
			nextStdoutAndStderr = true
			nextAppend = true
			continue
		}
		switch {
		case nextStdin:
			lastSession.Stdin = field
			nextStdin = false
		case nextStdout:
			lastSession.Stdout = field
			lastSession.StdoutAppend = nextAppend
			nextStdout = false
		case nextStderr:
			lastSession.Stderr = field
			lastSession.StderrAppend = nextAppend
			nextStderr = false
		case nextStdoutAndStderr:
			lastSession.Stdout = field
			lastSession.StdoutAppend = nextAppend
			lastSession.Stderr = field
			lastSession.StderrAppend = nextAppend
			nextStdoutAndStderr = false
		default:
			terms, err := parseBackquote(field)
			if err != nil {
				return nil, err
			}
			if len(terms) == 1 {
				lastSession.Fragments = append(lastSession.Fragments, Fragment{
					Term: field,
				})
			} else {
				f := Fragment{}
				for i := 0; i < len(terms); i += 2 {
					f.Texts = append(f.Texts, terms[i])
					if i+1 < len(terms) {
						subSessions, err := ParseCommandStr(terms[i+1])
						if err != nil {
							return nil, err
						} else if len(subSessions) > 1 || len(subSessions[0]) > 1 {
							return nil, ErrTooManySessionsInBackquote
						}
						f.Sessions = append(f.Sessions, subSessions[0][0])
					}
				}
				lastSession.Fragments = append(lastSession.Fragments, f)
			}
		}
	}
	if len(lastSession.Fragments) == 0 && lastField == "|" {
		return nil, ErrNoProcessAfterPipe
	}
	if nextStdin || nextStdout || nextStderr || nextStdoutAndStderr {
		return nil, fmt.Errorf("no target after '%s' %w", lastField, ErrNoRedirectTarget)
	}
	return result, nil
}
