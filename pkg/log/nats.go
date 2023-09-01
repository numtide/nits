package log

import (
	"bufio"
	"io"

	"github.com/nats-io/nats.go"
)

type NatsWriter struct {
	Conn    *nats.Conn
	Subject string

	reader *io.PipeReader
	writer *io.PipeWriter
}

func (w *NatsWriter) Close() (err error) {
	if w.writer != nil {
		err = w.writer.Close()
	}
	return
}

func (w *NatsWriter) Write(p []byte) (n int, err error) {
	if w.reader == nil {
		w.reader, w.writer = io.Pipe()
		go w.process()
	}
	return w.writer.Write(p)
}

func (w *NatsWriter) process() {
	scanner := bufio.NewScanner(w.reader)
	for scanner.Scan() {
		line := scanner.Text()
		msg := nats.NewMsg(w.Subject)
		msg.Data = []byte(line)
		if err := w.Conn.PublishMsg(msg); err != nil {
			println(err)
		}
	}
}
