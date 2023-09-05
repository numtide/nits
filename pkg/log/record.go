package log

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
	"github.com/numtide/nits/pkg/subject"
)

type Record struct {
	Level log.Level
	Msg   string
	Meta  map[string]string

	Subject   string
	AgentInfo *info.Response

	LoggedAt   time.Time
	ReceivedAt time.Time
}

func (r *Record) AgentSubject() string {
	return subject.AgentSubjectRegex().FindStringSubmatch(r.Subject)[1]
}

func (r *Record) Write(file *os.File) (n int, err error) {
	b := bytes.NewBuffer(nil)

	// todo handle errors
	// todo support multi ple formats
	b.WriteString(log.TimestampStyle.Render(r.ReceivedAt.Format(time.RFC3339)))
	b.WriteByte(' ')
	b.WriteString(levelStyle(r.Level).Render(r.Level.String()))
	b.WriteByte(' ')

	if r.AgentInfo != nil {
		b.WriteString(log.PrefixStyle.Render(r.AgentInfo.Name))
	} else {
		b.WriteString(log.PrefixStyle.Render(r.Subject))
	}

	b.WriteByte(' ')
	b.WriteString(log.MessageStyle.Render(r.Msg))
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

func Unmarshal(msg *nats.Msg, record *Record) (err error) {
	if msg == nil {
		return errors.New("msg cannot be nil")
	} else if record == nil {
		return errors.New("record cannot be nil")
	}

	var meta *nats.MsgMetadata
	if meta, err = msg.Metadata(); err != nil {
		return
	}

	record.Subject = msg.Subject
	record.ReceivedAt = meta.Timestamp
	record.Meta = make(map[string]string)

	dec := logfmt.NewDecoder(bytes.NewReader(msg.Data))

	for dec.ScanRecord() {
		for dec.ScanKeyval() {
			key := string(dec.Key())
			value := string(dec.Value())
			switch key {
			case "ts":
				if record.LoggedAt, err = time.Parse(time.RFC3339, value); err != nil {
					return
				}
			case "msg":
				record.Msg = value
			case "lvl":
				record.Level = log.ParseLevel(value)
			default:
				record.Meta[key] = value
			}
		}
	}

	return
}
