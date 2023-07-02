package agent

import (
	"context"
	"encoding/json"
	"net"
	"time"

	natshttp "github.com/brianmcgee/nats-http"

	"github.com/numtide/nits/pkg/nix"

	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/server"
	"github.com/numtide/nits/pkg/state"
	"golang.org/x/sync/errgroup"
)

func (a *Agent) listenForDeployment(ctx context.Context) error {
	deployment, err := a.js.KeyValue(state.DeploymentConfig.Bucket)
	if err != nil {
		return err
	}

	deploymentResult, err := a.js.KeyValue(state.DeploymentResultConfig.Bucket)
	if err != nil {
		return err
	}

	// listen for deployments using our nkey
	watch, err := deployment.Watch(a.nkey)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return watch.Stop()
		case entry, ok := <-watch.Updates():
			if !ok {
				// channel has been closed
				return nil
			}
			if entry == nil {
				// nothing available yet for our nkey
				continue
			}
			if entry.Operation() == nats.KeyValuePut {
				// only process puts
				var config server.Deployment
				if err = json.Unmarshal(entry.Value(), &config); err != nil {
					a.logger.Error("failed to unmarshal deployment update", "error", err)
					continue
				}
				a.onDeployment(&config, deploymentResult)
			}
		}
	}
}

func (a *Agent) onDeployment(deployment *server.Deployment, resultStore nats.KeyValue) {
	startedAt := time.Now()

	l := a.logger.New("action", deployment.Action, "closure", deployment.Closure)
	l.Info("checking current system closure")

	currentSystem, err := nix.CurrentSystemClosure()
	if err != nil {
		l.Error("failed to retrieve current system closure", "error", err)
	}

	if currentSystem == deployment.Closure {
		l.Info("current system matches deployment closure", "system", currentSystem)
		return
	}

	defer func() {
		elapsed := time.Since(startedAt)
		if err == nil {
			l.Info("end of deployment", "elapsed", elapsed)
		} else {
			l.Error("end of deployment", "elapsed", elapsed, "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	l.Info("initialising embedded binary cache")

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}

	cacheProxy := natshttp.Proxy{
		Subject: "nits.cache",
		Transport: &natshttp.Transport{
			Conn: a.conn.Conn,
		},
		Listener: listener,
	}

	eg.Go(func() error {
		if err := cacheProxy.Listen(ctx); err != nil {
			l.Error("cache proxy failed to listen", "error", err)
		}
		return err
	})

	eg.Go(func() (err error) {
		defer cancel()

		defer func() {
			// todo handle output that is larger than 1 MB and therefore too large for the KV store

			result := server.DeploymentResult{
				Deployment: *deployment,
				Error:      err,
			}

			b, err := json.Marshal(result)
			if err != nil {
				l.Error("failed to marshal deployment result to json", "error", err)
				return
			}

			_, err = resultStore.Put(a.nkey, b)
			if err != nil {
				l.Error("failed to write command output to object store", "error", err)
			}
		}()

		l.Info("copying closure from binary cache", "listenAddr", listener.Addr())

		err = nix.CopyFromBinaryCache(listener.Addr(), deployment.Closure, l)
		if err != nil {
			l.Error("failure whilst copying from binary cache")
			return err
		}

		// todo check if the agent binary has changed and perform a restart after switching
		l.Info("applying configuration")

		err = nix.SwitchToConfiguration(deployment, a.DryRun, l)
		if err != nil {
			l.Error("failed to apply configuration", "error", err)
			return err
		}

		return nil
	})

	err = eg.Wait()
}
