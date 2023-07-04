package state

import "github.com/nats-io/nats.go"

var AgentLogStreamConfig = &nats.StreamConfig{
	Name:     "agent-logs",
	Subjects: []string{"nits.agent.*.logs"},
}

var AgentDeploymentStreamConfig = &nats.StreamConfig{
	Name:              "agent-deployments",
	Subjects:          []string{"nits.agent.*.deployment"},
	MaxMsgsPerSubject: 1, // only retain the last
}

func InitStreams(js nats.JetStreamContext) (err error) {
	_, err = js.AddStream(AgentLogStreamConfig)
	if err != nil {
		return
	}
	_, err = js.AddStream(AgentDeploymentStreamConfig)
	return
}
