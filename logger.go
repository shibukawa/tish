package tish

import (
	"bytes"
	"io"
	"time"
)

type LogRoot struct {
	Logs   []Log  `json:"logs"`
	Option Option `json:"option"`
}

type LogLine struct {
	Epoch int64  `json:"epoch"`
	Line  string `json:"line"`
}

type Log struct {
	Epoch int64 `json:"start_at"`

	UserTime   time.Duration `json:"user_time"`
	SystemTime time.Duration `json:"system_time"`
	WallTime   time.Duration `json:"wall_time"`
	CPU        float64       `json:"cpu_usage"`
	ExitCode   int           `json:"exit_code"`

	Env map[string]string `json:"env"`

	Stdin  []LogLine `json:"stdin"`
	Stdout []LogLine `json:"stdout"`
	Stderr []LogLine `json:"stderr"`

	Command string `json:"command"`

	PreTasks   []Log `json:"pre_tasks"`
	ChildTasks []Log `json:"child_tasks"`
}

func DumpLog(w io.Writer) error {
	return nil
}

func DumpLogToString() (string, error) {
	var buffer bytes.Buffer
	err := DumpLog(&buffer)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
