package util

import (
	"io"
	"os"

	"github.com/juju/errors"
	"github.com/nats-io/nkeys"
	"golang.org/x/crypto/ssh"
)

func NewSigner(file *os.File) (ssh.Signer, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read key file")
	}
	return ssh.ParsePrivateKey(b)
}

func PublicKeyForSigner(signer ssh.Signer) (string, error) {
	key := signer.PublicKey()

	marshalled := key.Marshal()
	seed := marshalled[len(marshalled)-32:]

	encoded, err := nkeys.Encode(nkeys.PrefixByteUser, seed)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}
