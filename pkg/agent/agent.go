package agent

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"os/exec"

	log "github.com/inconshreveable/log15"

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

	nkey string
	conn *nats.Conn
	js   nats.JetStreamContext
}

func (a *Agent) Init() error {
	// test we can use nix
	cmd := exec.Command("nix", "run", "nixpkgs#hello")

	out, err := cmd.Output()
	_, _ = os.Stderr.Write(out)

	if err != nil {
		return err
	}

	// connect to nats server
	if err := a.connectNats(); err != nil {
		return err
	}

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

	conn, err := nats.Connect(nc.Url, natsOpts...)
	if err != nil {
		return errors.Annotatef(err, "nkey = %s", a.nkey)
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a jet stream context")
	}

	a.conn = conn
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
