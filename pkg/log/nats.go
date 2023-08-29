package log

import (
	"io"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsWriter struct {
	Conn        *nats.Conn
	Subject     string
	Delegate    io.Writer
	Interval    time.Duration
	PayloadSize int

	writeCh chan []byte
}

func (w *NatsWriter) Write(p []byte) (n int, err error) {
	if w.writeCh == nil {
		w.writeCh = make(chan []byte, 256)

		if w.PayloadSize == 0 {
			w.PayloadSize = 1024 * 512 // 512 kb
		}

		if w.Interval == 0 {
			w.Interval = 100 * time.Millisecond
		}

		go w.listen()
	}

	if w.Delegate == nil {
		w.writeCh <- p
		return len(p), nil
	}

	if n, err = w.Delegate.Write(p); err != nil {
		return
	}

	w.writeCh <- p[:n]

	return
}

func (w *NatsWriter) listen() {
	var buf []byte
	timeout := time.After(w.Interval)

	for {
		select {

		case <-timeout:
			w.publish(buf)
			buf = nil
			timeout = time.After(w.Interval)

		case b, ok := <-w.writeCh:
			if !ok {
				// channel has been closed
				break
			}
			buf = append(buf, b...)
			if len(buf) > w.PayloadSize {
				w.publish(buf)
				buf = nil
			}
		}
	}
}

func (w *NatsWriter) publish(b []byte) {
	if len(b) == 0 {
		return
	}

	msg := nats.NewMsg(w.Subject)
	msg.Data = b
	// swallow the error, publishing is best-effort for now
	_ = w.Conn.PublishMsg(msg)
}
