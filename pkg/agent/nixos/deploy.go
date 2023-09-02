package nixos

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/numtide/nits/pkg/agent/info"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ettle/strcase"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nuid"
	"github.com/nix-community/go-nix/pkg/nixpath"
	nlog "github.com/numtide/nits/pkg/log"
	"github.com/numtide/nits/pkg/nix"
	"github.com/numtide/nits/pkg/subject"
	"golang.org/x/sync/errgroup"
)

type DeployAction int

const (
	Switch DeployAction = iota
	Boot
	Test
	DryActivate
)

var deployErrGroup = errgroup.Group{}

type DeployRequest struct {
	Action  DeployAction `json:"action"`
	Closure string       `json:"closure"`
}

type DeployResponse struct {
	Id   string `json:"id"`
	Logs string `json:"logs"`
}

func onDeploy(req micro.Request) {
	var (
		err     error
		request DeployRequest
		closure *nixpath.NixPath
	)

	if err = json.Unmarshal(req.Data(), &request); err != nil {
		_ = req.Error("400", fmt.Sprintf("Failed to unmarshal request: %s", err), nil)
		return
	}

	if closure, err = nixpath.FromString(request.Closure); err != nil {
		_ = req.Error("400", fmt.Sprintf("Malformed closure: %s", err), nil)
		return
	}

	id := nuid.Next()
	logSubject := fmt.Sprintf("%s.NIXOS.DEPLOY.%s", subject.AgentLogs(NKey), id)

	scheduled := deployErrGroup.TryGo(func() (err error) {
		natsWriter := &nlog.NatsWriter{
			Conn:    Conn,
			Subject: logSubject,
		}

		l := log.NewWithOptions(io.MultiWriter(os.Stderr, natsWriter), log.Options{
			ReportTimestamp: true,
			Formatter:       log.LogfmtFormatter,
		})

		stdWriter := &nlog.Writer{Log: l.With("out", "std")}
		errWriter := &nlog.Writer{Log: l.With("out", "err")}

		ctx := context.Background()
		ctx = nix.SetStdOut(ctx, stdWriter)
		ctx = nix.SetStdError(ctx, errWriter)

		defer func() {
			_ = stdWriter.Close()
			_ = errWriter.Close()
			_ = natsWriter.Close()
		}()

		action := strcase.ToKebab(request.Action.String())
		l.Info("starting deployment", "action", action, "closure", closure)

		if err = nix.Build(closure, nil, ctx); err != nil {
			l.Error("failed to build closure", "error", err)
			return
		}

		if err = nix.Switch(closure, action, ctx); err != nil {
			l.Error("failed to switch configuration", "error", err)
			return
		}

		switch request.Action {
		case Boot, Switch:
			if err = nix.SetSystem(closure, ctx); err != nil {
				l.Error("failed to set system", "error", err)
				return
			}
		default:
			// do nothing
		}

		l.Info("deployment complete", "action", action, "closure", closure)

		return
	})

	if !scheduled {
		_ = req.Error("417", "A deployment is in progress.", nil)
		return
	}

	response := DeployResponse{
		Id:   id,
		Logs: logSubject,
	}

	if err = req.RespondJSON(response); err != nil {
		logger.Error("failed to respond", "error", err)
	}
	return
}

func Deploy(conn *nats.EncodedConn, nkey string, req DeployRequest) (resp *DeployResponse, err error) {
	err = conn.Request(subject.AgentService(nkey, "NIXOS.DEPLOY"), req, &resp, 10*time.Second)
	return
}

func DeployWithName(conn *nats.EncodedConn, name string, req DeployRequest) (resp *DeployResponse, err error) {
	var agentInfo info.Response
	if err = conn.Request(subject.AgentWithName(name), nil, &agentInfo, 10*time.Second); err != nil {
		return
	}
	return Deploy(conn, agentInfo.NKey, req)
}
