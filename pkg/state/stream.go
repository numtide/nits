package state

import "github.com/nats-io/nats.go"

var AgentLogStreamConfig = &nats.StreamConfig{
	Name:     "agent-logs",
	Subjects: []string{"nits.agent.*.logs"},
}

func InitStreams(js nats.JetStreamContext) (err error) {
	_, err = js.AddStream(AgentLogStreamConfig)
	return
}
