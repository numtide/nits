package cache

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"net/http"
)

const (
	DefaultNatsURL = "ns://127.0.0.1:4222"
)

var (
	DefaultCacheInfo = Info{
		StoreDir:      "/nix/store",
		WantMassQuery: true,
		Priority:      1,
	}
)

type Option func(opts *Options) error

type Options struct {
	NatsUrl string
	Info    Info
}

func NatsUrl(url string) Option {
	return func(opts *Options) error {
		opts.NatsUrl = url
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

	conn *nats.Conn
	js   nats.JetStreamContext

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

	conn, err := nats.Connect(s.Options.NatsUrl)
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
