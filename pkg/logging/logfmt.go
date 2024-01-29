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
	return RecordLogFmt
}

func (r *LogFmtRecord) Msg() *nats.Msg {
	return r.msg
}

func (r *LogFmtRecord) Write(file *os.File) (n int, err error) {
	b := bytes.NewBuffer(nil)

	// todo handle errors
	// todo support multiple formats
	styles := log.DefaultStyles()
	b.WriteString(styles.Timestamp.Render(r.Timestamp.Format(time.RFC3339)))
	b.WriteByte(' ')
	b.WriteString(levelStyle(r.Level).Render(r.Level.String()))
	b.WriteByte(' ')

	b.WriteString(styles.Prefix.Render(r.msg.Subject))

	b.WriteByte(' ')
	b.WriteString(styles.Message.Render(r.Text))
	b.WriteByte(' ')

	if r.AgentInfo != nil {
		b.WriteString(styles.Key.Render("nkey"))
		b.WriteByte('=')
		b.WriteString(r.AgentInfo.NKey)
		b.WriteByte(' ')
	}

	for k, v := range r.Meta {
		b.WriteString(styles.Key.Render(k))
		b.WriteByte('=')
		b.WriteString(styles.Value.Render(v))
		b.WriteByte(' ')
	}

	b.WriteByte('\n')
	return file.Write(b.Bytes())
}

func levelStyle(level log.Level) lipgloss.Style {
	return log.DefaultStyles().Levels[level]
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
				record.Level, err = log.ParseLevel(value)
				if err != nil {
					return err
				}
			default:
				record.Meta[key] = value
			}
		}
	}

	return
}
