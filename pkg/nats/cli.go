package nats

import (
	"crypto/rand"
	"os"

	"github.com/nats-io/jwt/v2"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"golang.org/x/crypto/ssh"
)

type CliOptions struct {
	Url             string `env:"NATS_URL" default:"nats://127.0.0.1:4222" help:"NATS server url."`
	JwtFile         string `type:"existingfile" env:"NATS_JWT_FILE" xor:"jwt"`
	Profile         string `env:"NATS_PROFILE" xor:"jwt,host,creds" help:"profile url in the form nsc://<OPERATOR>/<ACCOUNT>/<USER> e.g. nsc://Nits/Numtide/Admin"`
	HostKeyFile     string `type:"existingfile" env:"NATS_HOST_KEY_FILE" xor:"host"`
	CredentialsFile string `type:"existingfile" env:"NATS_CREDENTIALS_FILE" required:"" xor:"jwt,host,creds"`
}

func (n *CliOptions) ToNatsOptions() (opts []nats.Option, nkey string, claims *jwt.UserClaims, err error) {
	if n.Profile != "" {
		var kp nkeys.KeyPair
		var encodedJwt string

		if kp, encodedJwt, err = ReadProfile(n.Profile); err != nil {
			return
		}

		opts = append(opts, nats.UserJWT(
			func() (string, error) {
				return encodedJwt, nil
			}, func(nonce []byte) ([]byte, error) {
				defer kp.Wipe()
				sig, err := kp.Sign(nonce)
				return sig, err
			}))

		return
	} else if n.CredentialsFile != "" {
		var kp nkeys.KeyPair
		var encodedJwt string
		if kp, encodedJwt, err = ReadCredentials(n.CredentialsFile); err != nil {
			return
		}
		defer kp.Wipe()

		nkey, err = kp.PublicKey()
		opts = append(opts, nats.UserCredentials(n.CredentialsFile))

		claims, err = DecodeUserClaims(encodedJwt)
		return
	} else if n.JwtFile != "" {

		if claims, err = ReadUserClaims(n.JwtFile); err != nil {
			return
		}

		if n.HostKeyFile == "" {
			// we will authenticate with just a bearer JWT
			opts = append(opts, nats.UserCredentials(n.JwtFile))
		} else {
			// we will authenticate with a JWT and using the host's key file as the NKey
			var signer ssh.Signer
			if signer, err = NewSigner(n.HostKeyFile); err != nil {
				return
			} else if nkey, err = NKeyForSigner(signer); err != nil {
				return
			}

			opts = append(opts, nats.UserJWT(
				func() (jwt string, err error) {
					var b []byte
					if b, err = os.ReadFile(n.JwtFile); err != nil {
						return
					}
					return string(b), nil
				}, func(bytes []byte) ([]byte, error) {
					sig, err := signer.Sign(rand.Reader, bytes)
					if err != nil {
						return nil, err
					}
					return sig.Blob, err
				}))
		}
	} else {
		// this shouldn't be able to happen
		err = errors.New("unexpected state")
	}

	return
}
