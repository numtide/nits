package exec

import (
	"encoding/json"
	"os/exec"

	"github.com/charmbracelet/log"

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

	if err == nil {
		log.Debug("detected operator",
			"name", operator.Name,
			"serviceUrls", operator.OperatorServiceURLs,
			"accountServerUrl", operator.AccountServerURL,
		)
	}

	return
}
