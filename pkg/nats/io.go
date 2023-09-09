package nats

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/log"
	"github.com/juju/errors"

	"github.com/nats-io/nats.go"
)

const (
	EOS      = "EOS"
	EOSValue = "End-Of-Stream"
)

type EndOfStreamErr struct {
	Subject string
}

func (e EndOfStreamErr) Error() string {
	return fmt.Sprintf("end of stream: %s", e.Subject)
}

func IsEndOfStreamErr(err error) bool {
	var EOS EndOfStreamErr
	return errors.As(err, &EOS)
}

type Writer struct {
	Conn    *nats.Conn
	Subject string
	Headers nats.Header
}

func (w *Writer) newMsg() *nats.Msg {
	msg := nats.NewMsg(w.Subject)
	for key, values := range w.Headers {
		for _, value := range values {
			msg.Header.Add(key, value)
		}
	}
	return msg
}

func (w *Writer) Close() (err error) {
	msg := w.newMsg()
	msg.Header.Set(EOS, EOSValue)
	return w.Conn.PublishMsg(msg)
}

func (w *Writer) Write(p []byte) (n int, err error) {
	msg := w.newMsg()
	msg.Data = p
	n = len(p)
	if err = w.Conn.PublishMsg(msg); err != nil {
		log.Error("failed to publish message", "subject", w.Subject)
	}
	return
}

type Reader struct {
	Sub     *nats.Subscription
	Context context.Context

	reader *io.PipeReader
	writer *io.PipeWriter
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.Context == nil {
		r.Context = context.Background()
	}

	if r.reader == nil {
		r.reader, r.writer = io.Pipe()
		go r.processSubscription()
	}

	return r.reader.Read(p)
}

func (r *Reader) processSubscription() {
	var (
		n   int
		err error
		msg *nats.Msg
	)

	for {

		if msg, err = r.Sub.NextMsgWithContext(r.Context); err != nil {
			_ = r.writer.CloseWithError(err)
			return
		}

		if msg.Header.Get(EOS) == EOSValue {
			// nothing more to consume
			_ = r.writer.CloseWithError(&EndOfStreamErr{Subject: msg.Subject})
		}

		data := msg.Data

		for n, err = r.writer.Write(data); err == nil && n > 0; {
			if err != nil {
				_ = r.writer.CloseWithError(err)
			}
			if n == len(data) {
				break
			} else {
				data = data[n:]
			}
		}
	}
}

func IsEndOfStream(msg *nats.Msg) (result bool, err error) {
	if msg == nil {
		err = errors.New("msg cannot be nil")
	} else if msg.Header == nil {
		err = errors.New("msg.Header cannot be nil")
	} else {
		result = msg.Header.Get(EOS) == EOSValue
	}
	return
}
