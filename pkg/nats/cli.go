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

func (c *CliOptions) Connect() (conn *nats.Conn, err error) {
	var opts []nats.Option
	if opts, _, _, err = c.ToNatsOptions(); err != nil {
		return
	}
	return nats.Connect(c.Url, opts...)
}

func (c *CliOptions) ToNatsOptions() (opts []nats.Option, nkey string, claims *jwt.UserClaims, err error) {
	if c.Profile != "" {
		var kp nkeys.KeyPair
		var encodedJwt string

		if kp, encodedJwt, err = ReadProfile(c.Profile); err != nil {
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
	} else if c.CredentialsFile != "" {
		var kp nkeys.KeyPair
		var encodedJwt string
		if kp, encodedJwt, err = ReadCredentials(c.CredentialsFile); err != nil {
			return
		}
		defer kp.Wipe()

		nkey, err = kp.PublicKey()
		opts = append(opts, nats.UserCredentials(c.CredentialsFile))

		claims, err = DecodeUserClaims(encodedJwt)
		return
	} else if c.JwtFile != "" {

		if claims, err = ReadUserClaims(c.JwtFile); err != nil {
			return
		}

		if c.HostKeyFile == "" {
			// we will authenticate with just a bearer JWT
			opts = append(opts, nats.UserCredentials(c.JwtFile))
		} else {
			// we will authenticate with a JWT and using the host's key file as the NKey
			var signer ssh.Signer
			if signer, err = NewSigner(c.HostKeyFile); err != nil {
				return
			} else if nkey, err = NKeyForSigner(signer); err != nil {
				return
			}

			opts = append(opts, nats.UserJWT(
				func() (jwt string, err error) {
					var b []byte
					if b, err = os.ReadFile(c.JwtFile); err != nil {
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
