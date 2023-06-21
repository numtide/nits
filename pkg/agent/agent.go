package agent

import (
	"context"
	"fmt"

	log "github.com/inconshreveable/log15"

	"github.com/numtide/nits/pkg/util"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultInboxPrefix = "nits.inbox.agent"
	DefaultLogPrefix   = "nits.log.agent"
)

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

func LogSubjectPrefix(prefix string) Option {
	return func(opts *Options) error {
		if prefix == "" {
			return errors.New("log prefix cannot be empty")
		}
		opts.LogSubjectPrefix = prefix
		return nil
	}
}

type Options struct {
	NatsConfig       *config.Nats
	DryRun           bool
	LogSubjectPrefix string

	signer ssh.Signer
}

func GetDefaultOptions() Options {
	return Options{
		NatsConfig:       config.DefaultNatsConfig,
		DryRun:           false,
		LogSubjectPrefix: DefaultLogPrefix,
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
			Subject: fmt.Sprintf("%s.%s", a.Options.LogSubjectPrefix, a.nkey),
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

	// set logger for capturing nats errors
	nc.Logger = a.logger

	// customise the inbox prefix, appending the agent nkey
	if nc.InboxPrefix == "" {
		nc.InboxPrefix = DefaultInboxPrefix
	}
	nc.InboxPrefixFn = func(config *config.Nats, nkey string) string {
		return fmt.Sprintf("%s.%s", config.InboxPrefix, nkey)
	}

	// connect to nats
	conn, nkey, err := nc.ConnectNats()
	if err != nil {
		return errors.Annotatef(err, "nkey = "+nkey)
	}

	// get the jetstream context
	js, err := conn.JetStream()
	if err != nil {
		return err
	}

	// convert the connection to a json encoded connection
	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	// capture references
	a.js = js
	a.nkey = nkey
	a.conn = encoded

	a.logger.Info("connected to nats", "nkey", a.nkey)

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
