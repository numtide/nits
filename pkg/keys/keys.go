package keys

import (
	"fmt"
	"github.com/juju/errors"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"io"
	"os"
	"path/filepath"
)

type Set[T any] struct {
	KP     nkeys.KeyPair
	PubKey string
	Jwt    string
	Claims *T
}

func (s *Set[T]) EncodeClaims(kp nkeys.KeyPair) (err error) {
	switch v := interface{}(s.Claims).(type) {
	case *jwt.OperatorClaims:
		s.Jwt, err = v.Encode(kp)
	case *jwt.AccountClaims:
		s.Jwt, err = v.Encode(kp)
	case *jwt.UserClaims:
		s.Jwt, err = v.Encode(kp)
	default:
		return errors.New("unexpected type")
	}
	return
}

func (s *Set[T]) WriteJwt(dir string, name string) (err error) {
	if name == "" {
		name = s.PubKey + ".jwt"
	}
	return WriteJwt(s.Jwt, dir+"/"+name)
}

func (s *Set[T]) WriteCredentials(dir string, name string) (err error) {
	if name == "" {
		name = s.PubKey + ".creds"
	}
	return WriteCredentials(s.KP, s.Jwt, dir+"/"+name)
}

func newSet[T any](factory func() (nkeys.KeyPair, error)) (s Set[T], err error) {
	if s.KP, err = factory(); err != nil {
		return
	} else if s.PubKey, err = s.KP.PublicKey(); err != nil {
		return
	}
	return
}

func ReadJwt(path string) (jwt string, err error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return "", err
	}

	return string(b), nil
}

func WriteJwt(jwt string, path string) (err error) {
	if jwt == "" {
		return errors.New("jwt cannot be empty")
	}
	if err = os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		return err
	}

	var file *os.File
	if file, err = os.Create(path); err != nil {
		return err
	}

	defer func() {
		err = file.Close()
	}()

	_, err = io.WriteString(file, jwt)
	return
}

func ReadCredentials(path string) (kp nkeys.KeyPair, jwt string, err error) {
	if path == "" {
		return nil, "", errors.New("nits: Set.ReadCredentials path cannot be empty")
	}

	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return
	}

	if kp, err = nkeys.ParseDecoratedNKey(b); err != nil {
		return
	}

	if jwt, err = nkeys.ParseDecoratedJWT(b); err != nil {
		return
	}

	return
}

func WriteCredentials(kp nkeys.KeyPair, jwt string, path string) (err error) {
	if path == "" {
		return errors.New("path cannot be empty")
	}
	if kp == nil {
		return errors.New("kp cannot be nil")
	}

	if err = os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		return err
	}

	var seed []byte
	if seed, err = kp.Seed(); err != nil {
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

	if jwt != "" {
		str = fmt.Sprintf("-----BEGIN NATS USER JWT-----\n%s\n------END NATS USER JWT------\n\n", jwt)
		if _, err = io.WriteString(file, str); err != nil {
			return err
		}
	}

	str = fmt.Sprintf("-----BEGIN USER NKEY SEED-----\n%s\n------END USER NKEY SEED------", string(seed))
	_, err = io.WriteString(file, str)

	return err
}
