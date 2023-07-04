package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	natshttp "github.com/brianmcgee/nats-http"

	"github.com/numtide/nits/pkg/nix"

	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/server"
	"golang.org/x/sync/errgroup"
)

func (a *Agent) listenForDeployment(ctx context.Context) error {
	subject := fmt.Sprintf(a.SubjectPrefixFormat+".deployment", a.nkey)
	sub, err := a.js.SubscribeSync(subject, nats.DeliverLastPerSubject())
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return sub.Unsubscribe()
		default:
			msg, err := sub.NextMsg(1 * time.Second)
			if err == nats.ErrTimeout {
				// no currently available msg
				continue
			}

			if err != nil {
				a.log.Error("failed to retrieve next deployment msg", "error", err)
				continue
			}

			if msg == nil {
				continue
			}

			// we go ahead and ack the message because we don't want re-delivery in case of failure
			// instead a user must evaluate why it failed and publish a new deployment
			if err = msg.Ack(nats.AckWait(10 * time.Minute)); err != nil {
				a.log.Error("failed to ack deployment", "error", err)
				continue
			}

			var config server.Deployment
			if err = json.Unmarshal(msg.Data, &config); err != nil {
				a.log.Error("failed to unmarshal deployment", "error", err)
				continue
			}

			a.onDeployment(&config)
		}
	}
}

func (a *Agent) onDeployment(deployment *server.Deployment) {
	startedAt := time.Now()

	l := a.log.With("action", deployment.Action, "closure", deployment.Closure)
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
