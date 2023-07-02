package server

import (
	"context"
	"net"

	natshttp "github.com/brianmcgee/nats-http"
	"github.com/numtide/nits/pkg/services/cache"
	"golang.org/x/sync/errgroup"

	log "github.com/inconshreveable/log15"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
	"github.com/numtide/nits/pkg/state"
)

type Server struct {
	NatsConfig   *config.Nats
	CacheOptions *cache.Options
	CacheAddress string

	log  log.Logger
	conn *nats.Conn
}

func (s *Server) Run(ctx context.Context, log log.Logger) (err error) {
	// validate properties
	if s.NatsConfig == nil {
		return errors.New("server: Server.NatsConfig cannot be nil")
	}

	// create sub logger
	s.log = log.New("component", "server")

	// connect to nats
	if err = s.connectNats(); err != nil {
		return err
	}

	// create cache service
	c := cache.Cache{
		Conn:    s.conn,
		Options: *s.CacheOptions,
	}

	// create a http proxy for the cache service
	listener, err := net.Listen("tcp", s.CacheAddress)
	if err != nil {
		return err
	}

	proxy := natshttp.Proxy{
		Subject:  c.Options.Subject,
		Listener: listener,
		Transport: &natshttp.Transport{
			Conn: s.conn,
			// increase the subscription pending msg bytes to 512 MB
			PendingBytesLimit: 1024 * 1024 * 512,
		},
	}

	// run services in an error group
	eg := errgroup.Group{}

	eg.Go(func() error {
		return c.Listen(ctx, s.log)
	})

	eg.Go(func() error {
		return proxy.Listen(ctx)
	})

	err = eg.Wait()
	if err == context.Canceled {
		err = nil
	}

	return err
}

func (s *Server) connectNats() (err error) {
	var nkey string
	s.conn, nkey, err = s.NatsConfig.Connect(s.log)
	if err != nil {
		return errors.Annotatef(err, "nkey = "+nkey)
	}

	// initialise various stores and streams
	if err = state.Init(s.conn); err != nil {
		return err
	}

	return nil
}
