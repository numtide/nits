package config

import (
	"crypto/rand"
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/numtide/nits/pkg/util"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultNatsURL = "ns://127.0.0.1:4222"
)

var DefaultNatsConfig = &Nats{
	Url: DefaultNatsURL,
}

type Nats struct {
	Url         string
	Jwt         string
	Seed        string
	JwtFile     *os.File
	HostKeyFile *os.File
	InboxFormat string

	InboxPrefixFn func(config *Nats, nkey string) string
}

func (n Nats) Connect(log log.Logger, extra ...nats.Option) (conn *nats.Conn, nkey string, err error) {
	var opts []nats.Option
	if !(n.Seed == "" || n.Jwt == "") {
		opts = append(opts, nats.UserJWTAndSeed(n.Jwt, n.Seed))

		var keypair nkeys.KeyPair
		keypair, err = nkeys.FromSeed([]byte(n.Seed))
		if err != nil {
			return
		}

		nkey, err = keypair.PublicKey()
		if err != nil {
			return
		}
	}

	if n.HostKeyFile != nil {
		var signer ssh.Signer
		signer, err = util.NewSigner(n.HostKeyFile)
		if err != nil {
			return
		}

		nkey, err = util.PublicKeyForSigner(signer)
		if err != nil {
			return
		}

		opts = append(opts, nats.UserJWT(
			func() (string, error) {
				return n.Jwt, nil
			}, func(bytes []byte) ([]byte, error) {
				sig, err := signer.Sign(rand.Reader, bytes)
				if err != nil {
					return nil, err
				}
				return sig.Blob, err
			}))
	}

	if n.InboxPrefixFn != nil {
		opts = append(opts, nats.CustomInboxPrefix(n.InboxPrefixFn(&n, nkey)))
	} else if n.InboxFormat != "" {
		opts = append(opts, nats.CustomInboxPrefix(fmt.Sprintf(n.InboxFormat, nkey)))
	}

	// log nats errors
	opts = append(opts, nats.ErrorHandler(func(_ *nats.Conn, subscription *nats.Subscription, err error) {
		if err != nil {
			log.Error("nats error", "subscription", subscription, "error", err)
		}
	}))

	// append optional overrides and extra options
	opts = append(opts, extra...)

	conn, err = nats.Connect(n.Url, opts...)
	return
}
