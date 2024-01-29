package cli

import (
	"context"
	"os"
	"time"

	"github.com/numtide/nits/pkg/agent"
	"github.com/numtide/nits/pkg/agent/info"
	nlog "github.com/numtide/nits/pkg/logging"

	"github.com/numtide/nits/pkg/agent/nixos"

	"github.com/charmbracelet/log"
	"github.com/numtide/nits/internal/cmd"

	nnats "github.com/numtide/nits/pkg/nats"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
)

type agentDeployCmd struct {
	Nats nnats.CliOptions `embed:"" prefix:"nats-"`

	Action  string `enum:"switch,boot,test,dry-activate" default:"switch" help:"action to perform on the agent" `
	Closure string `arg:"" help:"store path of the NixOS closure to deploy"`

	Output bool   `help:"output agent's stdout and stderr"`
	Name   string `required:"" help:"the name given to the agent"`
}

func (d *agentDeployCmd) Run() error {
	if err := Cmd.Log.ConfigureLog(); err != nil {
		return err
	}

	return cmd.Run(func(ctx context.Context) (err error) {
		if _, err = os.Stat(d.Closure); os.IsNotExist(err) {
			return errors.New("store path does not exist")
		}

		var action nixos.DeployAction
		if action, err = nixos.DeployActionString(d.Action); err != nil {
			return
		}

		req := nixos.DeployRequest{
			Action:  action,
			Closure: d.Closure,
		}

		var (
			opts    []nats.Option
			js      nats.JetStreamContext
			conn    *nats.Conn
			encoded *nats.EncodedConn
			sub     *nats.Subscription
		)

		if opts, _, _, err = d.Nats.ToNatsOptions(); err != nil {
			return
		} else if conn, err = nats.Connect(d.Nats.Url, opts...); err != nil {
			return
		} else if encoded, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER); err != nil {
			return
		} else if js, err = conn.JetStream(); err != nil {
			return
		}

		log.Info("resolving agent", "name", d.Name)
		listCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		var (
			ok                bool
			target            *info.Response
			byName, bySubject map[string]*info.Response
		)

		if byName, err = agent.ListByFunc(listCtx, conn, func(response *info.Response) string {
			return response.Name
		}); err != nil {
			return
		} else if target, ok = byName[d.Name]; ok {
			log.Info("agent found", "name", d.Name, "nkey", target.NKey)
		} else {
			return errors.Errorf("could not find an agent named %s", d.Name)
		}

		bySubject = make(map[string]*info.Response)
		for _, v := range byName {
			bySubject[v.Subject] = v
		}

		var resp nixos.DeployResponse
		if resp, err = nixos.DeployWithContext(ctx, encoded, target.NKey, req); err != nil {
			return
		} else if sub, err = js.SubscribeSync(resp.Logs+".>", nats.DeliverAll(), nats.AckNone()); err != nil {
			return
		}

		log.Debug("listening for logs", "subject", resp.Logs)
		reader := nlog.RecordReader{Sub: sub, Context: ctx}

		var record nlog.Record
		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err = reader.Read()
				if errors.Is(err, nats.ErrTimeout) {
					err = nil
					continue
				} else if nnats.IsEndOfStreamErr(err) {
					var eos nnats.EndOfStreamErr
					errors.As(err, &eos)
					if eos.Subject == resp.Logs+".SYS" {
						err = nil
						return
					} else {
						continue
					}
				} else if err != nil {
					return
				}

				if !d.Output && record.Type() == nlog.RecordTerm {
					continue
				}

				_, _ = record.Write(os.Stderr)
			}
		}
	})
}
