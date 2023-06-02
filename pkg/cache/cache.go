package cache

import (
	"context"
	log "github.com/inconshreveable/log15"
	"io"
	"net/http"
	"os"

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
		NatsConfig: config.DefaultNatsConfig,
		Info:       DefaultCacheInfo,
	}
}

type Cache struct {
	Options Options

	log log.Logger

	conn *nats.Conn
	js   nats.JetStreamContext

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
	return nil
}

func (c *Cache) Run(ctx context.Context) (err error) {
	server := http.Server{
		Addr:    "localhost:3000",
		Handler: c.router,
	}

	go func() {
		<-ctx.Done()
		_ = server.Close()
	}()

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	return err
}

func (c *Cache) connectNats() error {
	var err error

	nc := c.Options.NatsConfig
	conn, err := nats.Connect(nc.Url, nats.UserJWTAndSeed(nc.Jwt, nc.Seed))
	if err != nil {
		return errors.Annotate(err, "failed to connect to NATS")
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a JetStream context")
	}

	nar, err := js.CreateObjectStore(&nats.ObjectStoreConfig{
		Bucket: "nar",
	})
	if err != nil {
		return errors.Annotate(err, "failed to create nar store")
	}

	narInfo, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "nar-info",
	})
	if err != nil {
		return errors.Annotate(err, "failed to create nar info store")
	}

	narInfoAccess, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "nar-info-access",
	})
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
