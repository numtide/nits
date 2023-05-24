package cache

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
	"go.uber.org/zap"
)

const (
	DefaultNatsURL = "ns://127.0.0.1:4222"
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

	NatsUrl  string
	NatsJwt  string
	NatsSeed string
}

func NatsConfig(url string, jwt string, seed string) Option {
	return func(opts *Options) error {
		opts.NatsUrl = url
		opts.NatsJwt = jwt
		opts.NatsSeed = seed
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
		NatsUrl: DefaultNatsURL,
		Info:    DefaultCacheInfo,
	}
}

type Cache struct {
	Options Options

	log *zap.Logger

	conn  *nats.Conn
	js    nats.JetStreamContext
	store nats.ObjectStore

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

	conn, err := nats.Connect(
		s.Options.NatsUrl,
		nats.UserJWTAndSeed(s.Options.NatsJwt, s.Options.NatsSeed),
	)
	if err != nil {
		return errors.Annotate(err, "failed to connect to NATS")
	}

	js, err := conn.JetStream()
	if err != nil {
		return errors.Annotate(err, "failed to create a JetStream context")
	}

	store, err := js.CreateObjectStore(&nats.ObjectStoreConfig{
		Bucket: "nix-cache",
	})
	if err != nil {
		return errors.Annotate(err, "failed to create nix-cache")
	}

	s.js = js
	s.conn = conn
	s.store = store

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
