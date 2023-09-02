package cli

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	nlog "github.com/numtide/nits/pkg/log"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/subject"
)

type agentLogsCmd struct {
	Nats nutil.CliOptions `embed:"nats-"`

	Since     *time.Duration `help:"Time ago from which to start replaying logs." default:"30s" xor:"start"`
	StartTime *time.Time     `help:"Time from which to start replaying logs." xor:"start"`

	Name string `arg:"" optional:""`
}

func (c *agentLogsCmd) Run() error {
	Cmd.Log.ConfigureLog()

	return cmd.Run(func(ctx context.Context) (err error) {
		var (
			conn *nats.Conn
			js   nats.JetStreamContext
			nkey string
			subj string
			sub  *nats.Subscription
			msg  *nats.Msg
		)

		subOpts := []nats.SubOpt{
			nats.AckNone(),
		}

		if c.StartTime != nil {
			subOpts = append(subOpts, nats.StartTime(*c.StartTime))
		} else if c.Since != nil {
			startTime := time.Now().Add(-(*c.Since))
			subOpts = append(subOpts, nats.StartTime(startTime))
		}

		if conn, err = c.Nats.Connect(); err != nil {
			return
		} else if js, err = conn.JetStream(); err != nil {
			return
		}

		if c.Name != "" {
			nkey, err = agent.ResolveNKey(ctx, conn, c.Name)
			if err != nil {
				return
			}
			subj = subject.AgentLogs(nkey) + ".>"
		} else {
			subj = subject.AgentLogsAll()
		}

		if sub, err = js.SubscribeSync(subj, subOpts...); err != nil {
			return
		}

		for {
			select {
			case <-ctx.Done():
				err = nil
				return
			default:
				if msg, err = sub.NextMsg(1 * time.Second); !(err == nil || errors.Is(err, nats.ErrTimeout)) {
					return
				}
				nlog.MsgToLog(log.Default(), msg)
			}
		}
	})
}
