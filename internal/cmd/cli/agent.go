package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/juju/errors"
	"github.com/numtide/nits/pkg/util"
	"golang.org/x/crypto/ssh"
)

type addAgentCmd struct {
	Cluster        string   `help:"Name of the account under which Agents will run"`
	Name           string   `help:"A name for the agent account"`
	PublicKey      string   `required:"" xor:"key"`
	PublicKeyFile  *os.File `required:"" xor:"key"`
	PrivateKeyFile *os.File `required:"" xor:"key"`
}

func (a *addAgentCmd) Run() (err error) {
	var nkey string

	if a.PrivateKeyFile != nil {
		var signer ssh.Signer
		if signer, err = util.NewSigner(a.PrivateKeyFile); err != nil {
			return errors.Annotate(err, "failed to parse private key file")
		}
		if nkey, err = util.NKeyForSigner(signer); err != nil {
			return err
		}
	} else {

		var pk ssh.PublicKey
		keyBytes := []byte(a.PublicKey)

		if !(a.PublicKey == "" || strings.Contains(a.PublicKey, "ssh-ed25519")) {
			keyBytes = []byte("ed25519 " + a.PublicKey)
		}

		if a.PublicKeyFile != nil {
			keyBytes, err = io.ReadAll(a.PublicKeyFile)
			if err != nil {
				return errors.Annotate(err, "failed to read public key file")
			}
		}

		if pk, _, _, _, err = ssh.ParseAuthorizedKey(keyBytes); err != nil {
			return errors.Annotate(err, "failed to parse public key")
		}

		nkey, err = util.NKeyForPublicKey(pk)
		if err != nil {
			return errors.Annotate(err, "failed to determine nkey for public key")
		}
	}

	return cmdSequence(
		// create a user for the agent
		nsc(
			"add", "user", "-a", a.Cluster,
			"-k", nkey,
			"-n", a.Name,
			"--allow-pub", "NITS.CACHE.>",
			"--allow-pubsub", fmt.Sprintf("NITS.AGENT.%s.>", nkey),
			"--allow-pub", "$JS.ACK.agent-deployments.>",
			"--allow-pub", "$JS.API.STREAM.NAMES",
			"--allow-pub", "$JS.API.CONSUMER.*.agent-deployments.>",
		),
	)
}
