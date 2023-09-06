package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type RequestError struct {
	Code        string
	Description string
	Data        []byte
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("status: %s err: %s", r.Code, r.Description)
}

func Request[Req any, Resp any](conn *nats.EncodedConn, subject string, req Req, resp *Resp, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RequestWithContext(ctx, conn, subject, req, resp)
}

func RequestWithContext[Req any, Resp any](ctx context.Context, conn *nats.EncodedConn, subject string, req Req, resp *Resp) (err error) {
	inbox := conn.Conn.NewRespInbox()

	var (
		msg *nats.Msg
		sub *nats.Subscription
	)

	if sub, err = conn.Conn.SubscribeSync(inbox); err != nil {
		return
	}
	defer func() {
		_ = sub.Unsubscribe()
	}()

	if err = conn.PublishRequest(subject, inbox, req); err != nil {
		return
	} else if msg, err = sub.NextMsgWithContext(ctx); err != nil {
		return
	}

	h := msg.Header
	if h.Get(micro.ErrorHeader) != "" {
		return &RequestError{
			Code:        h.Get(micro.ErrorCodeHeader),
			Description: h.Get(micro.ErrorHeader),
			Data:        msg.Data,
		}
	}

	return conn.Enc.Decode(msg.Subject, msg.Data, resp)
}
