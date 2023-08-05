package agent

import (
	"fmt"
	"os"

	"github.com/numtide/nits/pkg/util"
)

type nkeyCmd struct {
	KeyFile *os.File `arg:""`
}

func (a *nkeyCmd) Run() error {
	signer, err := util.NewSigner(Cmd.Nkey.KeyFile)
	if err != nil {
		return err
	}

	pub, err := util.NKeyForSigner(signer)
	if err != nil {
		return err
	}

	fmt.Println(pub)
	return nil
}
