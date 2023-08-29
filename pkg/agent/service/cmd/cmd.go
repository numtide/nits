package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
	"github.com/nats-io/nuid"
	"github.com/numtide/nits/pkg/agent/util"
	"github.com/numtide/nits/pkg/exec"
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

	logger = log.FromContext(ctx).With("service", "cmd")

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
	case ActionExecute:
		if response.Id, response.Logs, err = execute(request.Cmd); err != nil {
			_ = req.Error("500", fmt.Sprintf("Failed to execute command: %s", err), nil)
			return
		}

	case ActionCancel:
		// todo support cancelling a long running command

	case ActionUnknown:
		_ = req.Error("400", fmt.Sprintf("Action unknown: %s", request.Action), nil)
		return
	}

	if err = req.RespondJSON(response); err != nil {
		logger.Error("failed to respond", "error", err)
	}
}

func execute(cmd *Command) (id string, logSubject string, err error) {
	id = nuid.Next()
	logSubject = fmt.Sprintf("%s.CMD.%s", subject.AgentLogs(NKey), id)

	writer := nlog.NatsWriter{
		Conn:     Conn,
		Subject:  logSubject,
		Delegate: os.Stderr,
	}

	l := log.New(&writer)

	shellCmd := exec.ShellCmd{
		Name: cmd.Name,
		Args: cmd.Args,
	}

	shellCmd.Stdout = exec.Logger{Log: l.With("output", "stdout")}
	shellCmd.Stderr = exec.Logger{Log: l.With("output", "stderr")}

	go func() {
		if err := shellCmd.Exec(); err != nil {
			l.Error("failed to execute command", "error", err)
		}
	}()

	return
}
