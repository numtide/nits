package log

import (
	"bytes"
	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
	"os"
	"time"
)

type TerminalRecord struct {
	msg *nats.Msg
}

func (t *TerminalRecord) Type() RecordType {
	return RecordTypeTerminal
}

func (t *TerminalRecord) Msg() *nats.Msg {
	return t.msg
}

func (t *TerminalRecord) Write(file *os.File) (n int, err error) {
	if log.GetLevel() > log.InfoLevel {
		return
	}

	b := bytes.NewBuffer(nil)

	var meta *nats.MsgMetadata
	if meta, err = t.msg.Metadata(); err != nil {
		return
	}

	b.WriteString(log.TimestampStyle.Render(meta.Timestamp.Format(time.RFC3339)))
	b.WriteByte(' ')

	prefix := t.msg.Subject
	if name := GetAgentName(t.msg); name != "" {
		prefix = name
	}

	b.WriteString(log.PrefixStyle.Render(prefix + " | OUT"))
	b.WriteString("\n")
	b.WriteString(log.MessageStyle.Render(string(t.msg.Data)))
	b.WriteByte(' ')

	b.WriteString("\n")
	return file.Write(b.Bytes())
}

func UnmarshalTerminalRecord(msg *nats.Msg, record *TerminalRecord) (err error) {
	if msg.Header.Get(HeaderFormat) != HeaderFormatTerminal {
		return ErrUnexpectedFormat
	}
	record.msg = msg
	return
}
