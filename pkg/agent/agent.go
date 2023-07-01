package agent

import (
	"context"
	"fmt"
	"net/http"

	natshttp "github.com/brianmcgee/nats.http"

	log "github.com/inconshreveable/log15"

	"github.com/numtide/nits/pkg/util"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"golang.org/x/crypto/ssh"
)

const (
	DefaultInboxFormat      = "nits.agent.%s.inbox"
	DefaultLogSubjectFormat = "nits.agent.%s.logs"
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

func LogSubjectFormat(format string) Option {
	return func(opts *Options) error {
		if format == "" {
			return errors.New("log format cannot be empty")
		}
		opts.LogSubjectFormat = format
		return nil
	}
}

type Options struct {
	NatsConfig       *config.Nats
	DryRun           bool
	LogSubjectFormat string

	signer ssh.Signer
}

func GetDefaultOptions() Options {
	return Options{
		NatsConfig:       config.DefaultNatsConfig,
		DryRun:           false,
		LogSubjectFormat: DefaultLogSubjectFormat,
	}
}

type Agent struct {
	Options Options

	logger log.Logger

	nkey string
	conn *nats.EncodedConn
	js   nats.JetStreamContext

	cacheClient *http.Client
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
			Subject: fmt.Sprintf(a.Options.LogSubjectFormat, a.nkey),
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

	// customise the inbox prefix, appending the agent nkey
	if nc.InboxFormat == "" {
		nc.InboxFormat = DefaultInboxFormat
	}
	nc.InboxPrefixFn = func(config *config.Nats, nkey string) string {
		return fmt.Sprintf(config.InboxFormat, nkey)
	}

	// connect to nats
	conn, nkey, err := nc.Connect(a.logger)
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

	a.cacheClient = &http.Client{
		Transport: &natshttp.Transport{
			Conn: conn,
		},
	}

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
