package deploy

import (
	"context"
	"fmt"
	"net"

	"github.com/numtide/nits/pkg/config"

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
	ErrSystemUpToDate   = errors.ConstError("nits: current system matches new system")
	ErrSwitchComplete   = errors.ConstError("nits: switch to configuration has completed")
)

type NixosHandler struct {
	Conn             *nats.Conn
	NixStorePath     string
	CacheProxyConfig *config.CacheProxy
}

func (h *NixosHandler) init(log *log.Logger) error {
	log.Debug("initialise")

	if h.NixStorePath == "" {
		h.NixStorePath = DefaultNixStorePath
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

	// extra nix config to be passed when building the system closure using NIX_CONFIG env variable
	extraNixConfig := map[string]string{}

	// create a new error group with a derived context
	eg, ctx := errgroup.WithContext(ctx)

	if h.CacheProxyConfig != nil {

		l.Info("initialising cache proxy", "subject", h.CacheProxyConfig.Subject)

		var listener net.Listener
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return
		}

		cacheProxy := natshttp.Proxy{
			Subject: h.CacheProxyConfig.Subject,
			Transport: &natshttp.Transport{
				Conn: h.Conn,
			},
			Listener: listener,
		}

		eg.Go(func() error {
			return cacheProxy.Listen(ctx)
		})

		// configure nix to use the cache proxy
		extraNixConfig["extra-substituters"] = fmt.Sprintf("http://%s", listener.Addr())
		extraNixConfig["extra-trusted-public-keys"] = h.CacheProxyConfig.PublicKey
	}

	eg.Go(func() (err error) {
		l.Info("building system closure", "storePath", storePath)

		nixConfigStr := ""
		for k, v := range extraNixConfig {
			nixConfigStr = nixConfigStr + fmt.Sprintf("%s = %s\n", k, v)
		}

		var env []string
		if nixConfigStr != "" {
			env = append(env, "NIX_CONFIG="+nixConfigStr)
		}

		if err = nix.BuildSystemClosure(storePath, env, ctx); err != nil {
			l.Error("failed to build system closure", "error", err)
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
