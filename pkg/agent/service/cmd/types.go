package cmd

import (
	"time"
)

type Action int

const (
	Execute Action = iota
	Cancel
)

type Request struct {
	Action Action   `json:"action"`
	Id     string   `json:"id,omitempty"`
	Cmd    *Command `json:"cmd,omitempty"`
}

type Command struct {
	Name    string         `json:"name"`
	Args    []string       `json:"args"`
	Timeout *time.Duration `json:"timeout,omitempty"`
}

type Response struct {
	Action Action `json:"action"`
	Id     string `json:"id,omitempty"`
	Logs   string `json:"logs,omitempty"`
}
