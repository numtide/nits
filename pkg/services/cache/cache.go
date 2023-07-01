package cache

import (
	"context"
	natshttp "github.com/brianmcgee/nats.http"
	log "github.com/inconshreveable/log15"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
)

var DefaultCacheInfo = Info{
	StoreDir:      "/nix/store",
	WantMassQuery: true,
	Priority:      1,
}

type Options struct {
	Info      *Info
	SecretKey *signature.SecretKey

	Subject string
	Group   string
}

func (o Options) Validate() error {
	if o.Subject == "" {
		return errors.New("cache: Options.Subject cannot be nil")
	}

	if o.Info == nil {
		o.Info = &DefaultCacheInfo
	}

	if o.SecretKey == nil {
		return errors.New("cache: Options.SecretKey cannot be empty")
	}

	return nil
}

type Cache struct {
	Conn    *nats.Conn
	Options Options

	name   string
	logger log.Logger
}

func (c *Cache) Listen(ctx context.Context, log log.Logger) (err error) {
	// validate properties
	if c.Conn == nil {
		return errors.New("cache: Cache.Conn cannot be nil")
	}

	if err = c.Options.Validate(); err != nil {
		return err
	}

	// derive the cache 'name' from the signature public key
	c.name = c.Options.SecretKey.ToPublicKey().Name

	// create logger
	logOpts := []interface{}{"name", c.name, "subject", c.Options.Subject}
	if c.Options.Group != "" {
		logOpts = append(logOpts, "group", c.Options.Group)
	}

	c.logger = log.New(logOpts...)

	// create server
	srv := &natshttp.Server{
		Conn:    c.Conn,
		Subject: c.Options.Subject,
		Group:   c.Options.Group,
		Handler: c.createRouter(),
		ErrorHandler: func(err error) {
			c.logger.Error("natshttp.Server: error serving request", "error", err)
		},
	}

	return srv.Listen(ctx)
}
