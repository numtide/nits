package agent

import (
	"fmt"

	"github.com/numtide/nits/pkg/nutil"
)

type nkeyCmd struct {
	KeyFile string `arg:"" type:"existingfile"`
}

func (a *nkeyCmd) Run() error {
	signer, err := nutil.NewSigner(Cmd.Nkey.KeyFile)
	if err != nil {
		return err
	}

	pub, err := nutil.NKeyForSigner(signer)
	if err != nil {
		return err
	}

	fmt.Println(pub)
	return nil
}
