package cli

import (
	"bytes"
	"context"
	"os"
	"time"

	"github.com/numtide/nits/pkg/agent/nixos"

	"github.com/charmbracelet/log"
	"github.com/go-logfmt/logfmt"
	"github.com/numtide/nits/internal/cmd"

	nutil "github.com/numtide/nits/pkg/nats"

	"github.com/nats-io/nkeys"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
)

type deployCmd struct {
	Nats nutil.CliOptions `embed:""`

	Action  string `enum:"switch,boot,test,dry-activate" default:"switch" help:"action to perform on the agent" `
	Closure string `arg:"" help:"store path of the NixOS closure to deploy"`

	Nkey string `required:"" xor:"target" help:"nkey of the agent to target"`
	Name string `required:"" xor:"target" help:"the name given to the agent"`
}

func (d *deployCmd) Run() error {
	Cmd.Log.ConfigureLog()

	return cmd.Run(func(ctx context.Context) (err error) {
		if _, err = os.Stat(d.Closure); os.IsNotExist(err) {
			return errors.New("store path does not exist")
		}

		if d.Nkey != "" {
			// todo use kong to parse and validate
			if d.Nkey[0] != 'U' {
				return errors.New("nkey must start with a 'U'")
			} else if _, err = nkeys.FromPublicKey(d.Nkey); err != nil {
				return errors.Annotate(err, "invalid nkey")
			}
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
			msg     *nats.Msg
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

		var resp *nixos.DeployResponse
		if d.Nkey != "" {
			resp, err = nixos.Deploy(encoded, d.Nkey, req)
		} else {
			resp, err = nixos.DeployWithName(encoded, d.Name, req)
		}

		if err != nil {
			return
		}

		if sub, err = js.SubscribeSync(resp.Logs, nats.DeliverAll(), nats.AckNone()); err != nil {
			return err
		}

		defer func() {
			_ = sub.Unsubscribe()
		}()

		agentLog := log.WithPrefix("agent")

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if msg, err = sub.NextMsg(1 * time.Second); !(err == nil || errors.Is(err, nats.ErrTimeout)) {
					return
				}
				if msg != nil {

					dec := logfmt.NewDecoder(bytes.NewReader(msg.Data))
					for dec.ScanRecord() {

						var msg, lvl string
						var keyvals []interface{}

						for dec.ScanKeyval() {
							key := string(dec.Key())
							value := string(dec.Value())
							switch key {
							case "msg":
								msg = value
							case "lvl":
								lvl = value
							default:
								keyvals = append(keyvals, key)
								keyvals = append(keyvals, string(dec.Value()))
							}
						}

						switch lvl {
						case "warn":
							agentLog.Warn(msg, keyvals...)
						case "info":
							agentLog.Info(msg, keyvals...)
						case "debug":
							agentLog.Debug(msg, keyvals...)
						case "error", "fatal":
							agentLog.Error(msg, keyvals...)
							agentLog.Error(msg, keyvals...)
						default:
							agentLog.Info(msg, keyvals...)
						}
					}
				}
			}
		}
	})
}
