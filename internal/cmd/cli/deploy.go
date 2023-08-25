package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nats-io/nkeys"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/types"
)

type deployCmd struct {
	Context string `help:"NATS cli context to use when interacting with the NATS server"`
	Action  string `enum:"switch,boot,test,dry-activate" default:"switch" help:"action to perform on the agent" `
	Closure string `arg:"" help:"store path of the NixOS closure to deploy"`

	Nkey string `required:"" xor:"target" help:"nkey of the agent to target"`
	Name string `required:"" xor:"target" help:"the name given to the agent"`
}

func (d *deployCmd) Run() (err error) {
	if _, err = os.Stat(d.Closure); os.IsNotExist(err) {
		return errors.New("store path does not exist")
	}
	// todo replace static strings with a util that helps build them and ensure uniformity

	var subject string
	if d.Nkey != "" {

		// todo use kong to parse and validate
		if d.Nkey[0] != 'U' {
			return errors.New("nkey must start with a 'U'")
		} else if _, err = nkeys.FromPublicKey(d.Nkey); err != nil {
			return errors.Annotate(err, "invalid nkey")
		}

		subject = fmt.Sprintf("NITS.AGENT.%s.DEPLOYMENT", d.Nkey)
	} else if d.Name != "" {
		subject = fmt.Sprintf("NITS.AGENT.NAME.%s.DEPLOYMENT", d.Name)
	} else {
		return errors.New("one of nkey or name must be provided")
	}

	deployment := types.Deployment{
		Action:  types.ToDeployAction(d.Action),
		Closure: d.Closure,
	}

	var b []byte
	if b, err = json.Marshal(deployment); err != nil {
		return err
	}

	return cmdSequence(
		nats("--context", d.Context, "publish", subject, string(b)),
	)
}
