package cache

import (
	"context"
	"github.com/numtide/nits/pkg/state"
	"io"
	"net"
	"net/http"
	"os"

	log "github.com/inconshreveable/log15"

	"github.com/numtide/nits/pkg/config"

	"github.com/go-chi/chi/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
)

var DefaultCacheInfo = Info{
	StoreDir:      "/nix/store",
	WantMassQuery: true,
	Priority:      1,
}

type Option func(opts *Options) error

type Options struct {
	Name      string
	Info      Info
	SecretKey signature.SecretKey

	NatsConn   *nats.Conn
	NatsConfig *config.Nats

	BindAddress string
}

func BindAddress(address string) Option {
	return func(opts *Options) error {
		opts.BindAddress = address
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

func NatsConnection(conn *nats.Conn) Option {
	return func(opts *Options) error {
		if conn == nil {
			return errors.New("conn cannot be nil")
		}
		opts.NatsConn = conn
		return nil
	}
}

func PrivateKeyFile(file *os.File) Option {
	return func(opts *Options) error {
		bytes, err := io.ReadAll(file)
		if err != nil {
			return errors.Annotate(err, "failed to read private key file")
		}

		key, err := signature.LoadSecretKey(string(bytes))
		if err != nil {
			return errors.Annotate(err, "failed to load private key")
		}

		opts.Name = key.ToPublicKey().Name
		opts.SecretKey = key

		return nil
	}
}

func InfoConfig(storeDir string, wantMassQuery bool, priority int) Option {
	return func(opts *Options) error {
		// todo validation
		opts.Info = Info{
			StoreDir:      storeDir,
			WantMassQuery: wantMassQuery,
			Priority:      priority,
		}
		return nil
	}
}

func GetDefaultOptions() Options {
	return Options{
		BindAddress: "localhost:3000",
		NatsConfig:  config.DefaultNatsConfig,
		Info:        DefaultCacheInfo,
	}
}

type Cache struct {
	Options Options

	log log.Logger

	conn     *nats.Conn
	js       nats.JetStreamContext
	listener net.Listener

	nar           nats.ObjectStore
	narInfo       nats.KeyValue
	narInfoAccess nats.KeyValue

	router *chi.Mux
}

func (c *Cache) Init() (err error) {
	c.log.Info("init")
	defer func() {
		if err == nil {
			c.log.Info("init complete")
		} else {
			c.log.Error("init error", "error", err)
		}
	}()

	if err = c.connectNats(); err != nil {
		return err
	}

	c.createRouter()

	l, err := net.Listen("tcp", c.Options.BindAddress)
	if err != nil {
		return err
	}

	c.listener = l
	c.log.Info("listening", "addr", l.Addr())

	return nil
}

func (c *Cache) ListenAddr() (addr net.Addr) {
	if c.listener != nil {
		addr = c.listener.Addr()
	}
	return
}

func (c *Cache) Run(ctx context.Context) (err error) {
	logger := c.log.New("addr", c.ListenAddr())

	srv := http.Server{
		Handler: c.router,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
		_ = c.listener.Close()

		logger.Info("listening stopped")
	}()

	err = srv.Serve(c.listener)
	if err == http.ErrServerClosed {
		err = nil
	}
	return
}

func (c *Cache) connectNats() error {
	var err error
	var conn = c.Options.NatsConn

	if conn == nil {
		nc := c.Options.NatsConfig
		conn, err = nats.Connect(nc.Url, nats.UserJWTAndSeed(nc.Jwt, nc.Seed))
		if err != nil {
			return errors.Annotate(err, "failed to connect to NATS")
		}
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a JetStream context")
	}

	nar, err := state.Nar(js)
	if err != nil {
		return errors.Annotate(err, "failed to get nar object store")
	}

	narInfo, err := state.NarInfo(js)
	if err != nil {
		return errors.Annotate(err, "failed to create nar info kv store")
	}

	narInfoAccess, err := state.NarInfoAccess(js)
	if err != nil {
		return errors.Annotate(err, "failed to create nar info access store")
	}

	c.js = js
	c.conn = conn

	c.nar = nar
	c.narInfo = narInfo
	c.narInfoAccess = narInfoAccess

	return nil
}

func NewCache(logger log.Logger, options ...Option) (*Cache, error) {
	// process options
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	return &Cache{
		Options: opts,
		log:     logger,
	}, nil
}
