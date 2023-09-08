package log

import (
	"github.com/numtide/nits/pkg/agent/info"
	"github.com/numtide/nits/pkg/subject"
	"io"
	"time"

	"github.com/charmbracelet/log"

	"github.com/nats-io/nats.go"
)

const (
	HeaderEOF            = "EOF"
	HeaderEOFValue       = "End-Of-Stream"
	HeaderFormat         = "Format"
	HeaderFormatLogFmt   = "LogFmt"
	HeaderFormatTerminal = "Terminal"
	HeaderAgentName      = "AgentName"
	DefaultReadTimeout   = 1 * time.Second
)

type NatsWriter struct {
	Conn    *nats.Conn
	Subject string
	Headers nats.Header
}

func (w *NatsWriter) newMsg() *nats.Msg {
	msg := nats.NewMsg(w.Subject)
	for key, values := range w.Headers {
		for _, value := range values {
			msg.Header.Add(key, value)
		}
	}
	return msg
}

func (w *NatsWriter) Close() (err error) {
	msg := w.newMsg()
	msg.Header.Set(HeaderEOF, HeaderEOFValue)
	return w.Conn.PublishMsg(msg)
}

func (w *NatsWriter) Write(p []byte) (n int, err error) {
	msg := w.newMsg()
	msg.Data = p
	n = len(p)
	if err = w.Conn.PublishMsg(msg); err != nil {
		log.Error("failed to publish message", "subject", w.Subject)
	}
	return
}

type NatsReader struct {
	Sub     *nats.Subscription
	Timeout time.Duration

	reader *io.PipeReader
	writer *io.PipeWriter
}

func (r *NatsReader) Read(p []byte) (n int, err error) {
	if r.reader == nil {
		r.reader, r.writer = io.Pipe()
		go r.processSubscription()
	}
	if r.Timeout == 0 {
		r.Timeout = DefaultReadTimeout
	}

	return r.reader.Read(p)
}

func (r *NatsReader) processSubscription() {
	var (
		n   int
		err error
		msg *nats.Msg
	)

	for {

		if msg, err = r.Sub.NextMsg(r.Timeout); err != nil {
			_ = r.writer.CloseWithError(err)
			return
		}

		if msg.Header.Get(HeaderEOF) == HeaderEOFValue {
			// nothing more to consume
			_ = r.writer.Close()
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

type MsgProcessor func(msg *nats.Msg) error

func ProcessMsg(msg *nats.Msg, processors ...MsgProcessor) (err error) {
	for _, proc := range processors {
		if err = proc(msg); err != nil {
			return
		}
	}
	return
}

func GetAgentName(msg *nats.Msg) string {
	return msg.Header.Get(HeaderAgentName)
}

func ResolveAgentName(bySubject map[string]*info.Response) MsgProcessor {
	return func(msg *nats.Msg) error {
		if msg.Header == nil {
			msg.Header = make(nats.Header)
		}

		subject.AgentSubjectRegex()

		if agent, ok := bySubject[msg.Subject[:67]]; ok {
			msg.Header.Set(HeaderAgentName, agent.Name)
		}
		return nil
	}
}
