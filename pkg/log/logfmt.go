package log

import (
	"bytes"
	"io"
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

const (
	ErrUnexpectedFormat = errors.ConstError("unexpected format")
)

type FmtRecord struct {
	Level log.Level
	Msg   string
	Meta  map[string]string

	Subject   string
	AgentInfo *info.Response

	LoggedAt   time.Time
	ReceivedAt time.Time
}

func (r *FmtRecord) AgentSubject() string {
	return subject.AgentSubjectRegex().FindStringSubmatch(r.Subject)[1]
}

func (r *FmtRecord) Write(prefix string, file *os.File) (n int, err error) {
	b := bytes.NewBuffer(nil)

	// todo handle errors
	// todo support multiple formats
	b.WriteString(log.TimestampStyle.Render(r.ReceivedAt.Format(time.RFC3339)))
	b.WriteByte(' ')
	b.WriteString(levelStyle(r.Level).Render(r.Level.String()))
	b.WriteByte(' ')

	if prefix != "" {
		b.WriteString(log.PrefixStyle.Render(prefix))
	} else {
		b.WriteString(log.PrefixStyle.Render(prefix))
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

type FmtReader struct {
	Sub     *nats.Subscription
	Timeout time.Duration
}

func (r *FmtReader) Read() (record *FmtRecord, err error) {
	if r.Timeout == 0 {
		r.Timeout = DefaultReadTimeout
	}
	var msg *nats.Msg
	if msg, err = r.Sub.NextMsg(r.Timeout); err != nil {
		return
	}

	if msg.Header.Get(HeaderEOF) == HeaderEOFValue {
		return nil, io.EOF
	} else if msg.Header.Get(HeaderFormat) != HeaderFormatLogFmt {
		// skip this message
		return nil, ErrUnexpectedFormat
	}

	var meta *nats.MsgMetadata
	record = &FmtRecord{}

	if err = Unmarshal(msg, record); err != nil {
		return
	} else if meta, err = msg.Metadata(); err != nil {
		return
	}

	record.ReceivedAt = meta.Timestamp
	return
}

func Unmarshal(msg *nats.Msg, record *FmtRecord) (err error) {
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
