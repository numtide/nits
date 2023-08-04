package deploy

import (
	"context"
	"net"

	natshttp "github.com/brianmcgee/nats-http"
	"github.com/charmbracelet/log"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nix-community/go-nix/pkg/nixbase32"
	"github.com/numtide/nits/pkg/nix"
	"github.com/numtide/nits/pkg/types"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultNixStorePath = "/nix/store"
	DefaultCacheSubject = "NITS.CACHE"

	ErrSystemUpToDate = errors.ConstError("nits: current system matches new system")
	ErrSwitchComplete = errors.ConstError("nits: switch to configuration has completed")
)

type NixosHandler struct {
	Conn         *nats.Conn
	NixStorePath string
	CacheSubject string
}

func (h *NixosHandler) init(log *log.Logger) error {
	log.Debug("initialise")

	if h.NixStorePath == "" {
		h.NixStorePath = DefaultNixStorePath
	}

	if h.CacheSubject == "" {
		h.CacheSubject = DefaultCacheSubject
	}

	if h.Conn == nil {
		return errors.New("nits: NixosHandler.Conn cannot be nil")
	}

	log.Debug("initialise complete")
	return nil
}

func (h *NixosHandler) Apply(config *types.Deployment, ctx context.Context) (err error) {
	l := log.FromContext(ctx)
	if err = h.init(l); err != nil {
		return err
	}

	storePath, err := config.StorePath()
	if err != nil {
		return err
	}

	newHash := nixbase32.EncodeToString(storePath.Digest)
	l = log.With("hash", newHash)

	currentSystem, err := nix.CurrentSystemClosure()
	if err != nil {
		return err
	}

	currentHash := nixbase32.EncodeToString(currentSystem.Digest)
	if currentHash == newHash {
		return ErrSystemUpToDate
	}

	// create a new error group with a derived context
	eg, ctx := errgroup.WithContext(ctx)

	l.Info("initialising cache proxy")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}

	cacheProxy := natshttp.Proxy{
		Subject: h.CacheSubject,
		Transport: &natshttp.Transport{
			Conn: h.Conn,
		},
		Listener: listener,
	}

	eg.Go(func() error {
		return cacheProxy.Listen(ctx)
	})

	eg.Go(func() (err error) {
		l.Info("copying closure from binary cache", "listenAddr", listener.Addr())

		err = nix.CopyFromBinaryCache(listener.Addr(), config.Closure, ctx)
		if err != nil {
			return err
		}

		// todo check if the agent binary has changed and perform a restart after switching
		l.Info("switch to configuration")

		err = nix.SwitchToConfiguration(config, ctx)
		if err != nil {
			l.Error("failed to switch configuration", "error", err)
			return err
		}

		switch config.Action {
		case types.DeployActionBoot, types.DeployActionSwitch:
			err = nix.SetSystemProfile(storePath, ctx)
			if err != nil {
				l.Error("failed to set system profile", "error", err)
				return err
			}
		default:
			// do nothing
		}

		l.Info("switch to configuration complete")

		// we return an error to force the error group ctx to be cancelled
		return ErrSwitchComplete
	})

	err = eg.Wait()
	if err == ErrSwitchComplete {
		err = nil
	}

	return err
}
