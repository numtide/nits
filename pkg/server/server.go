package server

import (
	"context"

	"github.com/numtide/nits/pkg/cache"

	log "github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"github.com/numtide/nits/pkg/state"
)

const DefaultInboxFormat = "nits.server.%s.inbox"

type Option func(opts *Options) error

type InitFn func(srv *Server) error

type Options struct {
	NatsConfig   *config.Nats
	CacheOptions []cache.Option
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

func CacheOptions(options []cache.Option) Option {
	return func(opts *Options) error {
		opts.CacheOptions = options
		return nil
	}
}

func GetDefaultOptions() Options {
	return Options{}
}

type Server struct {
	Options Options
	logger  log.Logger

	conn *nats.EncodedConn
	js   nats.JetStreamContext

	cache *cache.Cache
}

func (s *Server) Init() (err error) {
	if err = s.connectNats(); err != nil {
		return err
	}

	if err = state.InitObjectStores(s.js); err != nil {
		return err
	}

	if err = state.InitKeyValueStores(s.js); err != nil {
		return err
	}

	if err = state.InitStreams(s.js); err != nil {
		return err
	}

	cacheOpts := s.Options.CacheOptions
	cacheOpts = append(cacheOpts, cache.NatsConnection(s.conn))

	c, err := cache.NewCache(
		s.logger.New("component", "cache"),
		cacheOpts...,
	)
	if err != nil {
		return err
	}

	if err = c.Init(); err != nil {
		return err
	}

	s.cache = c

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	return s.cache.Run(ctx)
}

func (s *Server) connectNats() error {
	nc := s.Options.NatsConfig

	inboxPrefix := nc.InboxFormat
	if inboxPrefix == "" {
		inboxPrefix = DefaultInboxFormat
	}

	nc.Logger = s.logger

	conn, nkey, err := nc.ConnectNats()
	if err != nil {
		return errors.Annotatef(err, "nkey = "+nkey)
	}

	encoded, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a jet stream context")
	}

	s.conn = encoded
	s.js = js

	return nil
}

func NewServer(logger log.Logger, options ...Option) (*Server, error) {
	// process options
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	return &Server{
		Options: opts,
		logger:  logger,
	}, nil
}
