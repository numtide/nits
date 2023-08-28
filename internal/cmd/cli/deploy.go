package cli

import (
	"encoding/json"
	"os"

	nutil "github.com/numtide/nits/pkg/nats"

	"github.com/numtide/nits/pkg/subject"

	"github.com/nats-io/nkeys"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/types"
)

type deployCmd struct {
	Nats nutil.CliOptions `embed:""`

	Action  string `enum:"switch,boot,test,dry-activate" default:"switch" help:"action to perform on the agent" `
	Closure string `arg:"" help:"store path of the NixOS closure to deploy"`

	Nkey string `required:"" xor:"target" help:"nkey of the agent to target"`
	Name string `required:"" xor:"target" help:"the name given to the agent"`
}

func (d *deployCmd) Run() (err error) {
	if _, err = os.Stat(d.Closure); os.IsNotExist(err) {
		return errors.New("store path does not exist")
	}

	var subj string
	if d.Nkey != "" {
		// todo use kong to parse and validate
		if d.Nkey[0] != 'U' {
			return errors.New("nkey must start with a 'U'")
		} else if _, err = nkeys.FromPublicKey(d.Nkey); err != nil {
			return errors.Annotate(err, "invalid nkey")
		}
		subj = subject.AgentDeploymentWithNKey(d.Nkey)
	} else if d.Name != "" {
		subj = subject.AgentDeploymentWithName(d.Name)
	} else {
		return errors.New("one of nkey or name must be provided")
	}

	deployment := types.Deployment{
		Action:  types.ToDeployAction(d.Action),
		Closure: d.Closure,
	}

	var (
		opts []nats.Option
		conn *nats.Conn
		js   nats.JetStream
	)

	if opts, _, _, err = d.Nats.ToNatsOptions(); err != nil {
		return
	} else if conn, err = nats.Connect(d.Nats.Url, opts...); err != nil {
		return
	} else if js, err = conn.JetStream(); err != nil {
		return
	}

	var data []byte
	if data, err = json.Marshal(deployment); err != nil {
		return
	}

	_, err = js.Publish(subj, data)

	return
}
