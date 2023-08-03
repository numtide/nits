package cache

import (
	"context"

	natshttp "github.com/brianmcgee/nats-http"
	"github.com/charmbracelet/log"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
	"github.com/numtide/nits/pkg/cache/state"
	"github.com/numtide/nits/pkg/config"

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
	Conn       *nats.Conn
	ConnConfig *config.Nats
	Options    Options

	log  *log.Logger
	name string
}

func (c *Cache) Listen(ctx context.Context, logger *log.Logger) (err error) {
	// validate properties
	if c.Conn == nil && c.ConnConfig == nil {
		return errors.New("cache: one of Cache.Conn or Cache.ConnConfig must be provided")
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

	c.log = logger.With(logOpts...)

	// connect to nats
	if err = c.connectNats(); err != nil {
		return err
	}

	// create server
	srv := &natshttp.Server{
		Conn:    c.Conn,
		Subject: c.Options.Subject,
		Group:   c.Options.Group,
		Handler: c.createRouter(),
		ErrorHandler: func(err error) {
			c.log.Error("natshttp.Server: error serving request", "error", err)
		},
	}

	return srv.Listen(ctx)
}

func (c *Cache) connectNats() (err error) {
	if c.Conn == nil {
		// create a connection based on the provided config
		var nkey string
		c.Conn, nkey, err = c.ConnConfig.Connect(c.log)
		if err != nil {
			return errors.Annotatef(err, "nkey = "+nkey)
		}
	}

	// initialise various stores and streams
	return state.Init(c.Conn)
}
