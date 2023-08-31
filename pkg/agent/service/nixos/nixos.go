package nixos

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/numtide/nits/pkg/agent/util"
	"github.com/numtide/nits/pkg/subject"
)

var (
	NKey string
	Conn *nats.Conn

	logger *log.Logger
)

func Init(ctx context.Context) (err error) {
	Conn = util.GetConn(ctx)
	NKey = util.GetNKey(ctx)

	logger = log.Default().With("service", "nixos")

	var srv micro.Service
	if srv, err = micro.AddService(Conn, micro.Config{
		Name:        "AgentNixos",
		Version:     "0.0.1",
		Description: "Nixos related functionality.",
	}); err != nil {
		return
	}

	group := srv.AddGroup(subject.AgentService(NKey, "NIXOS"))

	if err = group.AddEndpoint("INFO", micro.HandlerFunc(onInfo)); err != nil {
		return
	} else if err = group.AddEndpoint("DEPLOY", micro.HandlerFunc(onDeploy)); err != nil {
		return
	}

	deployErrGroup.SetLimit(1)

	return
}
