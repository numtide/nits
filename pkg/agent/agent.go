package agent

import (
	"context"
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
	"github.com/numtide/nits/pkg/agent/info"
	"github.com/numtide/nits/pkg/agent/nixos"

	"github.com/numtide/nits/pkg/agent/util"

	nutil "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/subject"

	"github.com/charmbracelet/log"
	nlog "github.com/numtide/nits/pkg/log"

	"github.com/nats-io/nats.go"
)

var (
	NatsOptions *nutil.CliOptions
	Conn        *nats.Conn
	NKey        string
	Claims      *jwt.UserClaims
)

func Run(ctx context.Context) (err error) {
	// connect to nats server
	if err = connectNats(); err != nil {
		return
	}

	// publish logs into nats
	writer := nlog.NatsWriter{
		Conn:    Conn,
		Subject: subject.AgentLogs(NKey) + ".SYS",
		Headers: nats.Header{
			nlog.HeaderFormat: []string{nlog.HeaderFormatLogFmt},
		},
	}
	defer func() {
		_ = writer.Close()
	}()

	log.SetOutput(io.MultiWriter(os.Stderr, &writer))

	// register services
	ctx = util.SetConn(ctx, Conn)
	ctx = util.SetNKey(ctx, NKey)
	ctx = util.SetClaims(ctx, Claims)

	log.Info("initialising services")
	if err = info.Init(ctx); err != nil {
		log.Errorf("failed to initialise info service", "error", err)
		return
	} else if err = nixos.Init(ctx); err != nil {
		log.Errorf("failed to initialise nixos service", "error", err)
		return
	}
	log.Info("services initialised")

	<-ctx.Done()
	return nil
}

func connectNats() (err error) {
	var opts []nats.Option

	if opts, NKey, Claims, err = NatsOptions.ToNatsOptions(); err != nil {
		return
	}
	opts = append(opts, nats.CustomInboxPrefix(subject.AgentInbox(NKey)))

	if Conn, err = nats.Connect(NatsOptions.Url, opts...); err != nil {
		return
	}
	log.Info("connected to nats", "nkey", NKey)

	return
}
