package cache

import (
	"context"

	nutil "github.com/numtide/nits/pkg/nats"

	natshttp "github.com/brianmcgee/nats-http"
	"github.com/charmbracelet/log"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/narinfo/signature"
	"github.com/numtide/nits/pkg/cache/state"
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
	Conn        *nats.Conn
	NatsOptions *nutil.CliOptions
	Options     Options

	log  *log.Logger
	name string
}

func (c *Cache) Listen(ctx context.Context, logger *log.Logger) (err error) {
	// validate properties
	if c.Conn == nil && c.NatsOptions == nil {
		return errors.New("cache: one of Cache.Conn or Cache.NatsOptions must be provided")
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

	msg := nats.NewMsg("")
	msg.Header.Get(nats.ACCOUNT_AUTHENTICATION_EXPIRED_ERR)

	return srv.Listen(ctx)
}

func (c *Cache) connectNats() (err error) {
	if c.Conn == nil {
		// create a connection based on the provided config
		var nkey string
		var opts []nats.Option
		if opts, nkey, _, err = c.NatsOptions.ToNatsOptions(); err != nil {
			return
		} else if c.Conn, err = nats.Connect(c.NatsOptions.Url, opts...); err != nil {
			err = errors.Annotate(err, "nkey = "+nkey)
			return
		}
	}

	// initialise various stores and streams
	return state.Init(c.Conn)
}
