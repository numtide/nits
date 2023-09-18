package nixos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/ettle/strcase"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nuid"
	"github.com/nix-community/go-nix/pkg/nixpath"
	nlog "github.com/numtide/nits/pkg/logging"
	nnats "github.com/numtide/nits/pkg/nats"
	"github.com/numtide/nits/pkg/nix"
	"github.com/numtide/nits/pkg/subject"
)

type DeployAction int

const (
	Switch DeployAction = iota
	Boot
	Test
	DryActivate
)

// the id of the deployment currently in progress
var currentDeployId = atomic.Value{}

type DeployRequest struct {
	Action  DeployAction `json:"action"`
	Closure string       `json:"closure"`
}

type DeployResponse struct {
	Id   string `json:"id"`
	Logs string `json:"logs"`
}

type DeployResult struct {
	Success bool `json:"success"`
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

	if !currentDeployId.CompareAndSwap("", id) {
		_ = req.Error("417", "A deployment is in progress.", nil)
		return
	}

	go func() {
		currentDeployId.Store(id)
		defer currentDeployId.Store("")

		logWriter := &nnats.Writer{
			Conn:    Conn,
			Subject: logSubject + ".SYS",
			Headers: nats.Header{
				nlog.HeaderFormat: []string{nlog.HeaderLogFmt},
			},
		}

		outWriter := &nnats.Writer{
			Conn:    Conn,
			Subject: logSubject + ".STDOUT",
			Headers: nats.Header{
				nlog.HeaderFormat: []string{nlog.HeaderTerm},
			},
		}

		errWriter := &nnats.Writer{
			Conn:    Conn,
			Subject: logSubject + ".STDERR",
			Headers: nats.Header{
				nlog.HeaderFormat: []string{nlog.HeaderTerm},
			},
		}

		l := log.New(io.MultiWriter(os.Stdout, logWriter))
		l.SetTimeFormat(time.RFC3339)
		l.SetLevel(log.DebugLevel)
		l.SetFormatter(log.LogfmtFormatter)
		l.SetReportTimestamp(true)

		ctx := context.Background()
		ctx = nix.SetStdOut(ctx, outWriter)
		ctx = nix.SetStdError(ctx, outWriter)

		defer func() {
			if err := errWriter.Close(); err != nil {
				log.Error("failed to close nats outWriter", "error", err)
			} else if err := outWriter.Close(); err != nil {
				log.Error("failed to close nats outWriter", "error", err)
			} else if err := logWriter.Close(); err != nil {
				log.Error("failed to close nats logWriter", "error", err)
			}
		}()

		action := strcase.ToKebab(request.Action.String())

		l.Info("starting deployment")

		l.Info("building closure", "closure", closure)
		if err = nix.Build(closure, nil, ctx); err != nil {
			l.Error("failed to build closure", "error", err)
			return
		}

		l.Info("switching configuration")
		if err = nix.Switch(closure, action, ctx); err != nil {
			l.Error("failed to switch configuration", "error", err)
			return
		}

		switch request.Action {
		case Boot, Switch:
			l.Info("setting system")
			if err = nix.SetSystem(closure, ctx); err != nil {
				l.Error("failed to set system", "error", err)
				return
			}
		default:
			// do nothing
		}

		l.Info("deployment complete")

		return
	}()

	response := DeployResponse{
		Id:   id,
		Logs: logSubject,
	}

	if err = req.RespondJSON(response); err != nil {
		logger.Error("failed to respond", "error", err)
	}
	return
}

func DeployWithContext(ctx context.Context, conn *nats.EncodedConn, nkey string, req DeployRequest) (resp DeployResponse, err error) {
	err = nnats.RequestWithContext(ctx, conn, subject.AgentService(nkey, "NIXOS.DEPLOY"), req, &resp)
	return
}
