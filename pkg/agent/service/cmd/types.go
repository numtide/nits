package cmd

import (
	"encoding/json"
	"time"
)

type Action int

const (
	ActionUnknown Action = iota
	ActionExecute
	ActionCancel
)

func (a Action) String() string {
	switch a {
	case ActionExecute:
		return "execute"
	case ActionCancel:
		return "cancel"
	}
	return "unknown"
}

func ToAction(s string) Action {
	switch s {
	case "execute":
		return ActionExecute
	case "cancel":
		return ActionCancel
	}
	return ActionUnknown
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Action) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*a = ToAction(str)
	return nil
}

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
