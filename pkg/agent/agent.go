package agent

import (
	"context"
	"crypto/rand"
	log "github.com/inconshreveable/log15"
	"io"
	"os"

	"github.com/numtide/nits/pkg/util"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"golang.org/x/crypto/ssh"
)

type Option func(opts *Options) error

func HostKeyFile(file *os.File) Option {
	return func(opts *Options) error {
		b, err := io.ReadAll(file)
		if err != nil {
			return errors.Annotate(err, "failed to read host key file")
		}

		signer, err := ssh.ParsePrivateKey(b)
		if err != nil {
			return errors.Annotate(err, "failed to parse host key file")
		}

		opts.signer = signer

		return nil
	}
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

type Options struct {
	NatsConfig *config.Nats

	signer ssh.Signer
}

func GetDefaultOptions() Options {
	return Options{
		NatsConfig: config.DefaultNatsConfig,
	}
}

type Agent struct {
	Options Options

	logger log.Logger

	conn *nats.Conn
	js   nats.JetStreamContext

	watcher nats.KeyWatcher
}

func (a *Agent) Init() error {
	if err := a.connectNats(); err != nil {
		return err
	}

	return nil
}

func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("listening for closures")
	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-a.watcher.Updates():
			if !ok {
				// update channel was closed
				return nil
			}
			if update == nil {
				// no value currently
				continue
			}
			if update.Operation() != nats.KeyValuePut {
				a.logger.Warn("unexpected op type received", "op", update.Operation())
				continue
			}
			a.logger.Info("Closure update received", "closure", string(update.Value()))
		}
	}
}

func (a *Agent) connectNats() error {
	nc := a.Options.NatsConfig

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
		a.logger.Info("loaded host key file", "publicKey", publicKey)

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

	closures, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "closures",
	})
	if err != nil {
		return errors.Annotate(err, "failed to create closures kv store")
	}

	// watch for changes based on our public key
	watcher, err := closures.Watch(publicKey)

	a.conn = conn
	a.js = js
	a.watcher = watcher

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
