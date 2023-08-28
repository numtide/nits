package nats

import (
	"crypto/rand"
	"os"

	"github.com/juju/errors"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"golang.org/x/crypto/ssh"
)

type NatsOptions struct {
	Url             string `name:"url" env:"NATS_URL" default:"nats://127.0.0.1:4222" help:"NATS server url."`
	JwtFile         string `name:"jwt-file" type:"existingfile" env:"NATS_JWT_FILE" xor:"jwt"`
	HostKeyFile     string `name:"host-key-file" type:"existingfile" env:"NATS_HOST_KEY_FILE" xor:"host"`
	CredentialsFile string `name:"credentials-file" type:"existingfile" env:"NATS_CREDENTIALS_FILE" required:"" xor:"jwt,host"`
}

func (n *NatsOptions) ToOpts() (opts []nats.Option, nkey string, claims *jwt.UserClaims, err error) {
	if n.CredentialsFile != "" {
		var kp nkeys.KeyPair
		if kp, claims, err = ReadUserCredentials(n.CredentialsFile); err != nil {
			return
		}
		nkey, err = kp.PublicKey()
		opts = append(opts, nats.UserCredentials(n.CredentialsFile))
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
