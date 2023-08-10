package state

import (
	"time"

	"github.com/nats-io/nats.go"
)

var AgentLogs *nats.StreamInfo

var AgentLogsConfig = nats.StreamConfig{
	Name: "agent-logs",
	Subjects: []string{
		"NITS.AGENT.*.LOGS",
	},
	MaxAge:  7 * 24 * time.Hour,
	Storage: nats.FileStorage,
}

var AgentDeployments *nats.StreamInfo

var AgentDeploymentsConfig = nats.StreamConfig{
	Name: "agent-deployments",
	Subjects: []string{
		"NITS.AGENT.*.DEPLOYMENT",
	},
	Storage:           nats.FileStorage,
	MaxMsgsPerSubject: 128,
}

func InitStreams(js nats.JetStreamContext) (err error) {
	if AgentLogs, err = js.AddStream(&AgentLogsConfig); err != nil {
		return err
	}

	AgentDeployments, err = js.AddStream(&AgentDeploymentsConfig)
	return err
}
