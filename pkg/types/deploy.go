package types

import (
	"encoding/json"

	"github.com/nix-community/go-nix/pkg/nixpath"
)

type DeployAction int

const (
	DeployActionUnknown DeployAction = iota
	DeployActionSwitch
	DeployActionBoot
	DeployActionTest
	DeployActionDryActivate
)

func (a DeployAction) String() string {
	switch a {
	case DeployActionSwitch:
		return "switch"
	case DeployActionBoot:
		return "boot"
	case DeployActionTest:
		return "test"
	case DeployActionDryActivate:
		return "dry-activate"
	}
	return "unknown"
}

func ToDeployAction(s string) DeployAction {
	switch s {
	case "switch":
		return DeployActionSwitch
	case "boot":
		return DeployActionBoot
	case "test":
		return DeployActionTest
	case "dry-activate":
		return DeployActionDryActivate
	}
	return DeployActionUnknown
}

func (a DeployAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *DeployAction) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*a = ToDeployAction(str)
	return nil
}

type Deployment struct {
	Action  DeployAction `json:"action"`
	Closure string       `json:"closure"`
}

func (d *Deployment) StorePath() (*nixpath.NixPath, error) {
	return nixpath.FromString(d.Closure)
}
