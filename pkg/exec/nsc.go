package exec

import (
	"encoding/json"
	"os/exec"

	"github.com/nats-io/nsc/v2/cmd"
)

func Nsc(args ...string) *exec.Cmd {
	return exec.Command("nsc", args...)
}

func DescribeOperator() (operator cmd.OperatorDescriber, err error) {
	var b []byte
	if b, err = Nsc("describe", "operator", "-J").Output(); err != nil {
		return
	}
	err = json.Unmarshal(b, &operator)
	return
}

func GenerateProfile(url string) (profile cmd.Profile, err error) {
	var b []byte
	if b, err = Nsc("generate", "profile", url).Output(); err != nil {
		return
	}
	err = json.Unmarshal(b, &profile)
	return
}
