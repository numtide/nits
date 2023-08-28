package agent

import (
	"context"
	"net/http"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/numtide/nits/pkg/nutil"
	"github.com/numtide/nits/pkg/subject"

	"github.com/numtide/nits/pkg/agent/deploy"

	"github.com/charmbracelet/log"
	nits_log "github.com/numtide/nits/pkg/log"

	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
)

type Agent struct {
	Deployer         deploy.Deployer
	NatsOptions      *nutil.NatsOptions
	CacheProxyConfig *config.CacheProxy

	log *log.Logger

	conn   *nats.Conn
	nkey   string
	claims *jwt.UserClaims

	cacheClient   *http.Client
	deployHandler deploy.Handler
}

func (a *Agent) Run(ctx context.Context, logger *log.Logger) error {
	a.log = logger

	// connect to nats server
	if err := a.connectNats(); err != nil {
		return err
	}

	// publish logs into nats
	writer := nits_log.NatsWriter{
		Conn:     a.conn,
		Subject:  subject.AgentLogs(a.nkey),
		Delegate: os.Stderr,
	}

	a.log.SetOutput(&writer)

	// set deploy handler
	switch a.Deployer {
	case deploy.DeployerNoOp:
		a.deployHandler = deploy.HandlerFunc(deploy.NoOpHandler)
	case deploy.DeployerNixos:
		a.deployHandler = &deploy.NixosHandler{
			Conn: a.conn,
		}
	}

	return a.listenForDeployment(ctx)
}

func (a *Agent) connectNats() (err error) {
	var opts []nats.Option
	if opts, a.nkey, a.claims, err = a.NatsOptions.ToOpts(); err != nil {
		return
	}

	opts = append(opts, nats.CustomInboxPrefix(subject.AgentInbox(a.nkey)))

	if a.conn, err = nats.Connect(a.NatsOptions.Url, opts...); err != nil {
		return
	}

	a.log.Info("connected to nats", "nkey", a.nkey)

	return
}
