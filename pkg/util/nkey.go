package util

import (
	"fmt"
	"github.com/nats-io/jwt/v2"
	"io"
	"os"
	"path/filepath"

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

type Keys struct {
	KP     nkeys.KeyPair
	PubKey string
	Jwt    string
}

func ReadOperatorJwt(path string) (k Keys, err error) {
	if k.Jwt, err = readJwt(path); err != nil {
		return
	}

	claims, err := jwt.DecodeOperatorClaims(k.Jwt)
	if err != nil {
		return
	}

	k.PubKey = claims.Subject
	return
}

func ReadAccountJwt(path string) (k Keys, err error) {
	if k.Jwt, err = readJwt(path); err != nil {
		return
	}

	claims, err := jwt.DecodeAccountClaims(k.Jwt)
	if err != nil {
		return
	}

	k.PubKey = claims.Subject
	return
}

func readJwt(path string) (jwt string, err error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return "", err
	}

	return string(b), nil
}

func (k *Keys) ReadCredentials(path string) (err error) {
	if path == "" {
		return errors.New("nits: Set.ReadCredentials path cannot be empty")
	}

	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return err
	}

	if k.KP, err = nkeys.ParseDecoratedNKey(b); err != nil {
		return err
	} else if k.PubKey, err = k.KP.PublicKey(); err != nil {
		return err
	}

	if k.Jwt, err = nkeys.ParseDecoratedJWT(b); err != nil {
		return err
	}

	return
}

func (k *Keys) WriteJwt(dir string) (err error) {
	if k.Jwt == "" {
		return errors.New("nits: Set.Jwt cannot be empty")
	}
	if k.PubKey == "" {
		return errors.New("nits: Set.PubKey cannot be empty")
	}
	if dir == "" {
		return errors.New("nits: Set.WriteJwt dir cannot be empty")
	}

	path := fmt.Sprintf("%s/%s.jwt", dir, k.PubKey)

	if err = os.MkdirAll(dir, 0744); err != nil {
		return err
	}

	var file *os.File
	if file, err = os.Create(path); err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	_, err = io.WriteString(file, k.Jwt)
	return
}

func (k *Keys) WriteCredentials(path string) (err error) {
	if path == "" {
		return errors.New("nits: Set.WriteCredentials path cannot be empty")
	}
	if k.KP == nil {
		return errors.New("nits: Set.KP cannot be nil")
	}
	if k.PubKey == "" {
		return errors.New("nits: Set.PubKey cannot be empty")
	}

	if err = os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		return err
	}

	var seed []byte
	if seed, err = k.KP.Seed(); err != nil {
		return err
	}

	var str string
	var file *os.File
	if file, err = os.Create(path); err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	if k.Jwt != "" {
		str = fmt.Sprintf("-----BEGIN NATS USER JWT-----\n%s\n------END NATS USER JWT------\n\n", k.Jwt)
		if _, err = io.WriteString(file, str); err != nil {
			return err
		}
	}

	str = fmt.Sprintf("-----BEGIN USER NKEY SEED-----\n%s\n------END USER NKEY SEED------", string(seed))
	_, err = io.WriteString(file, str)

	return err
}

func NewKeys(fn func() (nkeys.KeyPair, error)) (*Keys, error) {
	kp, err := fn()
	if err != nil {
		return nil, err
	}
	pubKey, err := kp.PublicKey()
	if err != nil {
		return nil, err
	}
	return &Keys{
		KP:     kp,
		PubKey: pubKey,
	}, nil
}
