package agent

import (
	"fmt"
	"github.com/numtide/nits/pkg/util"
	"os"
)

type nkeyCmd struct {
	KeyFile *os.File `arg:""`
}

func (a *nkeyCmd) Run() error {
	signer, err := util.NewSigner(Cmd.Nkey.KeyFile)
	if err != nil {
		return err
	}

	pub, err := util.PublicKeyForSigner(signer)
	if err != nil {
		return err
	}

	fmt.Println(pub)
	return nil
}
