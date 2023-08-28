package nats

import (
	"io"
	"os"

	"github.com/juju/errors"
	"github.com/nats-io/nkeys"
	"golang.org/x/crypto/ssh"
)

func NewSigner(path string) (signer ssh.Signer, err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read key file")
	}
	return ssh.ParsePrivateKey(b)
}

func NKeyForSigner(signer ssh.Signer) (string, error) {
	return NKeyForPublicKey(signer.PublicKey())
}

func NKeyForPublicKey(pk ssh.PublicKey) (string, error) {
	marshalled := pk.Marshal()
	seed := marshalled[len(marshalled)-32:]

	encoded, err := nkeys.Encode(nkeys.PrefixByteUser, seed)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func NKeyForPublicKeyFile(file *os.File) (string, error) {
	b, err := io.ReadAll(file)
	if err != nil {
		return "", errors.Annotate(err, "failed to read key file")
	}

	var pk ssh.PublicKey
	if pk, err = ssh.ParsePublicKey(b); err != nil {
		return "", err
	}
	return NKeyForPublicKey(pk)
}
