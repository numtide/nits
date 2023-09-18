package logging

import (
	"bytes"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/go-logfmt/logfmt"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/numtide/nits/pkg/agent/info"
)

type LogFmtRecord struct {
	Level log.Level
	Text  string
	Meta  map[string]string

	AgentInfo *info.Response

	Timestamp time.Time

	msg *nats.Msg
}

func (r *LogFmtRecord) Type() RecordType {
	return RecordTypeLogFmt
}

func (r *LogFmtRecord) Msg() *nats.Msg {
	return r.msg
}

func (r *LogFmtRecord) Write(file *os.File) (n int, err error) {
	b := bytes.NewBuffer(nil)

	// todo handle errors
	// todo support multiple formats
	b.WriteString(log.TimestampStyle.Render(r.Timestamp.Format(time.RFC3339)))
	b.WriteByte(' ')
	b.WriteString(levelStyle(r.Level).Render(r.Level.String()))
	b.WriteByte(' ')

	prefix := r.msg.Subject
	if name := GetAgentName(r.msg); name != "" {
		prefix = name
	}

	b.WriteString(log.PrefixStyle.Render(prefix))

	b.WriteByte(' ')
	b.WriteString(log.MessageStyle.Render(r.Text))
	b.WriteByte(' ')

	if r.AgentInfo != nil {
		b.WriteString(log.KeyStyle.Render("nkey"))
		b.WriteByte('=')
		b.WriteString(r.AgentInfo.NKey)
		b.WriteByte(' ')
	}

	for k, v := range r.Meta {
		b.WriteString(log.KeyStyle.Render(k))
		b.WriteByte('=')
		b.WriteString(log.ValueStyle.Render(v))
		b.WriteByte(' ')
	}

	b.WriteByte('\n')
	return file.Write(b.Bytes())
}

func levelStyle(level log.Level) lipgloss.Style {
	switch level {
	case log.DebugLevel:
		return log.DebugLevelStyle
	case log.InfoLevel:
		return log.InfoLevelStyle
	case log.WarnLevel:
		return log.WarnLevelStyle
	case log.ErrorLevel:
		return log.ErrorLevelStyle
	case log.FatalLevel:
		return log.FatalLevelStyle
	default:
		return lipgloss.NewStyle()
	}
}

func UnmarshalLogFmtRecord(msg *nats.Msg, record *LogFmtRecord) (err error) {
	if msg == nil {
		return errors.New("msg cannot be nil")
	} else if record == nil {
		return errors.New("record cannot be nil")
	}

	record.msg = msg
	record.Meta = make(map[string]string)

	dec := logfmt.NewDecoder(bytes.NewReader(msg.Data))

	for dec.ScanRecord() {
		for dec.ScanKeyval() {
			key := string(dec.Key())
			value := string(dec.Value())
			switch key {
			case "ts":
				if record.Timestamp, err = time.Parse(time.RFC3339, value); err != nil {
					return
				}
			case "msg":
				record.Text = value
			case "lvl":
				record.Level = log.ParseLevel(value)
			default:
				record.Meta[key] = value
			}
		}
	}

	return
}
