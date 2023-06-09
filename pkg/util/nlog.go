package util

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"github.com/nats-io/nats.go"
	"time"
)

type NatsLogger struct {
	Js      nats.JetStreamContext
	Subject string
}

func (nl *NatsLogger) Log(r *log15.Record) error {
	msg := nats.NewMsg(nl.Subject)
	h := msg.Header

	h.Set(r.KeyNames.Time, r.Time.Format(time.RFC3339))
	h.Set(r.KeyNames.Lvl, r.Lvl.String())
	h.Set("call", r.Call.String())

	for i := 0; i < len(r.Ctx); i += 2 {
		h.Set(r.Ctx[i].(string), fmt.Sprintf("%v", r.Ctx[i+1]))
	}

	msg.Data = []byte(r.Msg)

	// fire and forget
	_, err := nl.Js.PublishMsgAsync(msg)
	return err
}
