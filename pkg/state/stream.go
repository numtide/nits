package state

import "github.com/nats-io/nats.go"

var LogStreamConfig = &nats.StreamConfig{
	Name:     "logs",
	Subjects: []string{"nits.log.>"},
}

func InitStreams(js nats.JetStreamContext) (err error) {
	_, err = js.AddStream(LogStreamConfig)
	return
}
