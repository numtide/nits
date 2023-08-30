package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/numtide/nits/pkg/nix"
	"io"
	"os"
	"os/exec"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nuid"
	"github.com/numtide/nits/pkg/agent/util"
	nlog "github.com/numtide/nits/pkg/log"
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

	logger = log.Default().With("service", "cmd")
	logger.SetFormatter(log.LogfmtFormatter)

	_, err = micro.AddService(Conn, micro.Config{
		Name:        "AgentCmd",
		Version:     "0.0.1",
		Description: "Execute commands on the host machine.",
		Endpoint: &micro.EndpointConfig{
			Subject: subject.AgentService(NKey, "CMD"),
			Handler: micro.HandlerFunc(handler),
		},
	})

	return
}

func handler(req micro.Request) {
	var err error
	var request Request
	var response Response

	logger.Debug("handling request", "data", string(req.Data()))

	if err = json.Unmarshal(req.Data(), &request); err != nil {
		_ = req.Error("500", fmt.Sprintf("Failed to unmarshal request: %s", err), req.Data())
		return
	}

	response.Action = request.Action

	switch request.Action {
	case Execute:
		if response.Id, response.Logs, err = execute(request.Cmd); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to execute command: %s", err), nil)
			return
		}

	case Cancel:
	// todo support cancelling a long running command

	default:
		_ = req.Error("400", fmt.Sprintf("Unknown action: %s", request.Action.String()), nil)
	}

	if err = req.RespondJSON(response); err != nil {
		logger.Error("failed to respond", "error", err)
	}
}

func execute(cmd *Command) (id string, logSubject string, err error) {
	id = nuid.Next()
	logSubject = fmt.Sprintf("%s.CMD.%s", subject.AgentLogs(NKey), id)

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

	c := exec.Command(cmd.Name, cmd.Args...)
	// forward output into NATS
	c.Stdout = io.MultiWriter(os.Stdout, stdWriter)
	c.Stderr = io.MultiWriter(os.Stderr, errWriter)

	go func() {
		defer func() {
			_ = stdWriter.Close()
			_ = errWriter.Close()
			_ = natsWriter.Close()
		}()

		if err := c.Run(); err != nil {
			logger.Error("failed to run command", "error", err)
		}
	}()

	return
}
