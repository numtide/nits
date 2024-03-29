package cli

import (
	"context"
	"os"
	"time"

	"github.com/charmbracelet/log"

	"github.com/numtide/nits/pkg/agent/info"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/internal/cmd"
	"github.com/numtide/nits/pkg/agent"
	nlog "github.com/numtide/nits/pkg/logging"
	nnats "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/subject"
)

type agentLogs struct {
	Nats nnats.CliOptions `embed:"" prefix:"nats-"`

	Since     *time.Duration `help:"Time ago from which to start replaying logs." default:"5m" xor:"start"`
	StartTime *time.Time     `help:"Time from which to start replaying logs." xor:"start"`

	Output bool   `help:"output agent's stdout and stderr"`
	Name   string `arg:"" optional:""`
}

func (c *agentLogs) Run() error {
	if err := Cmd.Log.ConfigureLog(); err != nil {
		return err
	}

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

		// get a list of agents and index the responses

		var agents []*info.Response
		if agents, err = agent.List(listCtx, conn); err != nil {
			return
		}

		var byName map[string]*info.Response
		if byName, err = agent.IndexByName(agents); err != nil {
			return
		}

		var byNKey map[string]*info.Response
		if byNKey, err = agent.IndexByNKey(agents); err != nil {
			return
		}

		// set agent indices in the context for the log writer
		ctx = nlog.SetAgentsByName(ctx, byName)
		ctx = nlog.SetAgentsByNKey(ctx, byNKey)

		// decide whether we are listening for a specific agents logs or all agents

		if c.Name != "" {
			if agentInfo, ok := byName[c.Name]; ok {
				subj = subject.AgentLogs(agentInfo.NKey) + ".>"
			} else {
				return errors.Errorf("could not find an agent with name = %s", c.Name)
			}
		} else {
			subj = subject.AgentLogsAll()
		}

		// start the subscription

		if sub, err = js.SubscribeSync(subj, subOpts...); err != nil {
			return
		}

		log.Debug("listening for logs", "subject", subj)

		// start reading the log records

		var record nlog.Record
		reader := nlog.RecordReader{Sub: sub, Context: ctx}

		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err = reader.Read()
				if nnats.IsEndOfStreamErr(err) || errors.Is(err, nats.ErrTimeout) {
					err = nil
					continue
				} else if err != nil {
					return
				}

				if !c.Output && record.Type() == nlog.RecordTerm {
					continue
				}

				_, _ = record.Write(os.Stderr)
			}
		}
	})
}
