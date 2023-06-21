package agent

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/nats-io/nkeys"

	log "github.com/inconshreveable/log15"

	"github.com/numtide/nits/pkg/util"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"golang.org/x/crypto/ssh"
)

const DefaultInboxPrefix = "nits.inbox.agent"

type Option func(opts *Options) error

func NatsConfig(config *config.Nats) Option {
	return func(opts *Options) error {
		if config == nil {
			return errors.New("config cannot be nil")
		}
		opts.NatsConfig = config
		return nil
	}
}

func SwitchDryRun(dryRun bool) Option {
	return func(opts *Options) error {
		opts.DryRun = dryRun
		return nil
	}
}

type Options struct {
	NatsConfig *config.Nats
	DryRun     bool

	signer ssh.Signer
}

func GetDefaultOptions() Options {
	return Options{
		NatsConfig: config.DefaultNatsConfig,
		DryRun:     false,
	}
}

type Agent struct {
	Options Options

	logger log.Logger

	nkey string
	conn *nats.EncodedConn
	js   nats.JetStreamContext
}

func (a *Agent) Init() error {
	// connect to nats server
	if err := a.connectNats(); err != nil {
		return err
	}

	multiHandler := log.MultiHandler(
		a.logger.GetHandler(),
		&util.NatsLogger{
			Js:      a.js,
			Subject: "nits.log.agent." + a.nkey,
		},
	)

	// mixin nats logging
	a.logger.SetHandler(multiHandler)

	return nil
}

func (a *Agent) Run(ctx context.Context) error {
	return a.listenForDeployment(ctx)
}

func (a *Agent) connectNats() error {
	nc := a.Options.NatsConfig

	var natsOpts []nats.Option
	if nc.Seed != "" {
		natsOpts = append(natsOpts, nats.UserJWTAndSeed(nc.Jwt, nc.Seed))
		keypair, err := nkeys.FromSeed([]byte(nc.Seed))
		if err != nil {
			return err
		}
		nkey, err := keypair.PublicKey()
		if err != nil {
			return err
		}
		a.nkey = nkey
	}

	if nc.HostKeyFile != nil {

		signer, err := util.NewSigner(nc.HostKeyFile)
		if err != nil {
			return err
		}

		nkey, err := util.PublicKeyForSigner(signer)
		if err != nil {
			return err
		}

		a.logger.Info("loaded host key file", "nkey", nkey)

		natsOpts = append(natsOpts, nats.UserJWT(
			func() (string, error) {
				return nc.Jwt, nil
			}, func(bytes []byte) ([]byte, error) {
				sig, err := signer.Sign(rand.Reader, bytes)
				if err != nil {
					return nil, err
				}
				return sig.Blob, err
			}))

		a.nkey = nkey
	}

	// set a custom inbox prefix by appending nkey to configured prefix
	inboxPrefix := nc.InboxPrefix
	if inboxPrefix == "" {
		inboxPrefix = DefaultInboxPrefix
	}

	natsOpts = append(natsOpts, nats.CustomInboxPrefix(fmt.Sprintf("%s.%s", inboxPrefix, a.nkey)))

	// capture nats errors in the logging
	natsOpts = append(natsOpts, nats.ErrorHandler(func(_ *nats.Conn, subscription *nats.Subscription, err error) {
		if err != nil {
			a.logger.Error("nats error", "subscription", subscription, "error", err)
		}
	}))

	conn, err := nats.Connect(nc.Url, natsOpts...)
	if err != nil {
		return errors.Annotatef(err, "nkey = %s", a.nkey)
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a jet stream context")
	}

	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	a.conn = encoded
	a.js = js

	return nil
}

func NewAgent(logger log.Logger, options ...Option) (*Agent, error) {
	// process options
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	return &Agent{
		Options: opts,
		logger:  logger,
	}, nil
}
