package cache

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/numtide/nits/pkg/config"

	"github.com/go-chi/chi/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
	"go.uber.org/zap"
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

	log *zap.Logger

	conn *nats.Conn
	js   nats.JetStreamContext

	nar     nats.ObjectStore
	narInfo nats.KeyValue

	router *chi.Mux
}

func (s *Cache) Init() (err error) {
	s.log.Info("init")
	defer func() {
		if err == nil {
			s.log.Info("init complete")
		} else {
			s.log.Error("init error", zap.Error(err))
		}
	}()

	if err = s.connectNats(); err != nil {
		return err
	}

	s.createRouter()
	return nil
}

func (s *Cache) Run(ctx context.Context) (err error) {
	server := http.Server{
		Addr:    "localhost:3000",
		Handler: s.router,
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

func (s *Cache) connectNats() error {
	var err error

	c := s.Options.NatsConfig
	conn, err := nats.Connect(c.Url, nats.UserJWTAndSeed(c.Jwt, c.Seed))
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

	s.js = js
	s.conn = conn
	s.nar = nar
	s.narInfo = narInfo

	return nil
}

func NewCache(log *zap.Logger, options ...Option) (*Cache, error) {
	// process options
	opts := GetDefaultOptions()
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	return &Cache{
		Options: opts,
		log:     log,
	}, nil
}
