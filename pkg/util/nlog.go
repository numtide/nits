package util

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/nats-io/nats.go"
	"time"
)

type NatsLogger struct {
	Conn    *nats.Conn
	Subject string
}

func (nl *NatsLogger) Log(r *log15.Record) error {
	msg := nats.NewMsg(nl.Subject)
	h := msg.Header

	h.Set("time", r.Time.Format(time.RFC3339))
	h.Set("level", r.Lvl.String())

	for i := 0; i < len(r.Ctx); i += 2 {
		h.Set(r.Ctx[i].(string), fmt.Sprintf("%v", r.Ctx[i+1]))
	}
	// todo stack

	msg.Data = []byte(r.Msg)

	return nl.Conn.PublishMsg(msg)
}
