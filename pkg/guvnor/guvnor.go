package guvnor

import (
	"context"
	"crypto/rand"
	log "github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"github.com/numtide/nits/pkg/state"
	"github.com/numtide/nits/pkg/util"
)

type Option func(opts *Options) error

type InitFn func(srv *Guvnor) error

type Options struct {
	NatsConfig *config.Nats
}

func NatsConfig(config *config.Nats) Option {
	return func(opts *Options) error {
		if config == nil {
			return errors.New("config cannot be nil")
		}
		opts.NatsConfig = config
		return nil
	}
}

func GetDefaultOptions() Options {
	return Options{}
}

type Guvnor struct {
	Options Options
	logger  log.Logger

	conn *nats.Conn
	js   nats.JetStreamContext
}

func (g *Guvnor) Init() error {
	if err := g.connectNats(); err != nil {
		return err
	}

	if err := state.InitObjectStores(g.js); err != nil {
		return err
	}

	return state.InitKeyValueStores(g.js)
}

func (g *Guvnor) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (g *Guvnor) connectNats() error {
	nc := g.Options.NatsConfig

	var natsOpts []nats.Option
	if nc.Seed != "" {
		natsOpts = append(natsOpts, nats.UserJWTAndSeed(nc.Jwt, nc.Seed))
	}

	var publicKey string
	if nc.HostKeyFile != nil {

		signer, err := util.NewSigner(nc.HostKeyFile)
		if err != nil {
			return err
		}

		publicKey, err = util.PublicKeyForSigner(signer)
		g.logger.Info("loaded host key file", "publicKey", publicKey)

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
	}

	conn, err := nats.Connect(nc.Url, natsOpts...)
	if err != nil {
		return errors.Annotatef(err, "nkey = %s", publicKey)
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a jet stream context")
	}

	g.conn = conn
	g.js = js

	return nil
}

func NewGuvnor(logger log.Logger, options ...Option) (*Guvnor, error) {
	// process options
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	return &Guvnor{
		Options: opts,
		logger:  logger,
	}, nil
}
