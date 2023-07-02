package agent

import (
	"context"
	"fmt"
	"net/http"

	natshttp "github.com/brianmcgee/nats-http"

	log "github.com/inconshreveable/log15"

	"github.com/numtide/nits/pkg/util"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/config"
)

type Agent struct {
	DryRun              bool
	NatsConfig          *config.Nats
	SubjectPrefixFormat string

	logger log.Logger

	nkey string
	conn *nats.EncodedConn
	js   nats.JetStreamContext

	cacheClient *http.Client
}

func (a *Agent) Run(ctx context.Context, logger log.Logger) error {
	a.logger = logger.New("component", "agent")

	// connect to nats server
	if err := a.connectNats(); err != nil {
		return err
	}

	multiHandler := log.MultiHandler(
		logger.GetHandler(),
		&util.NatsLogger{
			Js:      a.js,
			Subject: fmt.Sprintf(a.SubjectPrefixFormat+".logs", a.nkey),
		},
	)

	// mixin nats logging
	a.logger.SetHandler(multiHandler)

	return a.listenForDeployment(ctx)
}

func (a *Agent) connectNats() (err error) {
	nc := a.NatsConfig

	// customise the inbox prefix, appending the agent nkey
	if nc.InboxFormat == "" {
		nc.InboxFormat = a.SubjectPrefixFormat + ".inbox"
	}
	nc.InboxPrefixFn = func(config *config.Nats, nkey string) string {
		return fmt.Sprintf(config.InboxFormat, nkey)
	}

	// connect to nats
	conn, nkey, err := nc.Connect(a.logger)
	if err != nil {
		return errors.Annotatef(err, "nkey = "+nkey)
	}

	a.nkey = nkey

	// get the jetstream context
	a.js, err = conn.JetStream()
	if err != nil {
		return err
	}

	// convert the connection to a json encoded connection
	a.conn, err = nats.NewEncodedConn(conn, nats.JSON_ENCODER)
	if err != nil {
		return err
	}

	a.cacheClient = &http.Client{
		Transport: &natshttp.Transport{
			Conn:              conn,
			PendingBytesLimit: 1024 * 1024 * 512,
		},
	}

	a.logger.Info("connected to nats", "nkey", a.nkey)

	return nil
}
