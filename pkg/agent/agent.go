package agent

import (
	"context"
	"os"

	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/subject"

	"github.com/numtide/nits/pkg/agent/deploy"

	"github.com/charmbracelet/log"
	nlog "github.com/numtide/nits/pkg/log"

	"github.com/nats-io/nats.go"
)

var (
	Log         = log.Default()
	Deployer    = deploy.DeployerNoOp
	NatsOptions *nutil.CliOptions
	Conn        *nats.Conn
	NKey        string

	deployHandler deploy.Handler
)

func Run(ctx context.Context) (err error) {

	// connect to nats server
	if err = connectNats(); err != nil {
		return
	}

	// publish logs into nats
	if Log == nil {
		Log = log.Default()
	}

	writer := nlog.NatsWriter{
		Conn:     Conn,
		Subject:  subject.AgentLogs(NKey),
		Delegate: os.Stderr,
	}

	Log.SetOutput(&writer)

	// set deploy handler
	switch Deployer {
	case deploy.DeployerNoOp:
		deployHandler = deploy.HandlerFunc(deploy.NoOpHandler)
	case deploy.DeployerNixos:
		deployHandler = &deploy.NixosHandler{
			Conn: Conn,
		}
	}

	return listenForDeployment(ctx)
}

func connectNats() (err error) {
	var opts []nats.Option
	if opts, NKey, _, err = NatsOptions.ToNatsOptions(); err != nil {
		return
	}

	opts = append(opts, nats.CustomInboxPrefix(subject.AgentInbox(NKey)))

	if Conn, err = nats.Connect(NatsOptions.Url, opts...); err != nil {
		return
	}

	Log.Info("connected to nats", "nkey", NKey)

	return
}
