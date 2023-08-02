package server

import (
	"context"
	natshttp "github.com/brianmcgee/nats-http"
	"github.com/charmbracelet/log"
	"github.com/juju/errors"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/auth"
	"github.com/numtide/nits/pkg/config"
	"github.com/numtide/nits/pkg/services/cache"
	"github.com/numtide/nits/pkg/state"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
)

type Server struct {
	DataDir      string
	ClientConfig *config.NatsClient
	ServerConfig string

	CacheOptions *cache.Options
	CacheAddress string

	log *log.Logger
	srv *server.Server

	user *auth.Set[jwt.UserClaims]
	conn *nats.Conn
}

func (s *Server) Run(ctx context.Context, logger *log.Logger) (err error) {
	if s.DataDir == "" {
		if s.DataDir, err = os.Getwd(); err != nil {
			return err
		}
	}

	s.log = logger
	s.log.SetPrefix("nits")

	// run embedded nats server
	natsLog := logger.With()
	natsLog.SetPrefix("nats")

	if err = s.runNats(ctx, logger); err != nil {
		return err
	}

	if s.user, err = auth.ReadUserCredentials(s.DataDir + "/creds/nits-server.creds"); err != nil {
		return
	} else if err = s.connectNats(); err != nil {
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
	if s.srv != nil {
		// override with embedded client url
		s.ClientConfig = &config.NatsClient{
			Url: s.srv.ClientURL(),
			Jwt: s.user.Jwt,
		}
		var seed []byte
		if seed, err = s.user.KP.Seed(); err != nil {
			return err
		}
		s.ClientConfig.Seed = string(seed)
	}

	var nkey string
	s.conn, nkey, err = s.ClientConfig.Connect(s.log)
	if err != nil {
		return errors.Annotatef(err, "nkey = "+nkey)
	}

	// initialise various stores and streams
	if err = state.Init(s.conn); err != nil {
		return err
	}

	return nil
}
