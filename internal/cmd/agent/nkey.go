package agent

import (
	"fmt"

	"github.com/numtide/nits/pkg/nats"
)

type nkeyCmd struct {
	KeyFile string `arg:"" type:"existingfile"`
}

func (a *nkeyCmd) Run() error {
	signer, err := nats.NewSigner(Cmd.Nkey.KeyFile)
	if err != nil {
		return err
	}

	pub, err := nats.NKeyForSigner(signer)
	if err != nil {
		return err
	}

	fmt.Println(pub)
	return nil
}
