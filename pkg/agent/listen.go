package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/numtide/nits/pkg/types"

	"github.com/nats-io/nats.go"
)

func (a *Agent) listenForDeployment(ctx context.Context) error {
	subject := fmt.Sprintf(a.SubjectPrefixFormat+".DEPLOYMENT", a.nkey)

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
			if err = msg.Ack(); err != nil {
				a.log.Error("failed to ack deployment", "error", err)
				continue
			}

			var config types.Deployment
			if err = json.Unmarshal(msg.Data, &config); err != nil {
				a.log.Error("failed to unmarshal deployment", "error", err)
				continue
			}

			startedAt := time.Now()
			l := a.log.With()

			l.Info("new deployment", "action", config.Action, "closure", config.Closure)

			err = a.deployHandler.Apply(&config, log.WithContext(ctx, l))
			elapsed := time.Now().Sub(startedAt)

			if err != nil {
				l.Error("deployment failed", "error", err, "elapsed", elapsed)
			} else {
				l.Info("deployment complete", "elapsed", elapsed)
			}
		}
	}
}
