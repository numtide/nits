package cli

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/numtide/nits/pkg/agent/info"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	nlog "github.com/numtide/nits/pkg/log"
	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/subject"
)

type agentLogsCmd struct {
	Nats nutil.CliOptions `embed:"" prefix:"nats-"`

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
			subj string
			sub  *nats.Subscription
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

		listCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		var agents []*info.Response
		if agents, err = agent.List(listCtx, conn); err != nil {
			return
		}

		var byName, bySubject map[string]*info.Response

		if byName, err = agent.IndexByName(agents); err != nil {
			return err
		} else if bySubject, err = agent.IndexBySubject(agents); err != nil {
			return err
		}

		if c.Name != "" {
			if agentInfo, ok := byName[c.Name]; ok {
				subj = subject.AgentLogs(agentInfo.NKey) + ".>"
			} else {
				return errors.Errorf("could not find an agent with name = %s")
			}
		} else {
			subj = subject.AgentLogsAll()
		}

		if sub, err = js.SubscribeSync(subj, subOpts...); err != nil {
			return
		}

		reader := nlog.FmtReader{Sub: sub}

		var record *nlog.FmtRecord

		for {
			record, err = reader.Read()
			if errors.Is(err, io.EOF) || errors.Is(err, nlog.ErrUnexpectedFormat) {
				err = nil
				continue
			} else if err != nil {
				return
			}

			prefix := record.AgentSubject()
			if target, ok := bySubject[record.AgentSubject()]; ok {
				prefix = target.Name
			}

			_, _ = record.Write(prefix, os.Stderr)
		}
	})
}
