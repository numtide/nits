package logging

import (
	"context"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	nnats "github.com/numtide/nits/pkg/nats"
)

const (
	ErrUnexpectedFormat = errors.ConstError("unexpected format")

	HeaderFormat = "Fmt"
	HeaderLogFmt = "LogFmt"
	HeaderTerm   = "Term"
)

type RecordReader struct {
	Sub     *nats.Subscription
	Context context.Context
}

func (r *RecordReader) Read() (record Record, err error) {
	var (
		msg   *nats.Msg
		isEOS bool
	)
	if msg, err = r.Sub.NextMsgWithContext(r.Context); err != nil {
		return
	} else if isEOS, err = nnats.IsEndOfStream(msg); err != nil {
		return
	} else if isEOS {
		return nil, nnats.EndOfStreamErr{Subject: r.Sub.Subject}
	}

	switch msg.Header.Get(HeaderFormat) {
	case HeaderTerm:
		tRecord := &TerminalRecord{}
		if err = UnmarshalTerminalRecord(r.Context, msg, tRecord); err == nil {
			record = tRecord
		}
	case HeaderLogFmt:
		lfRecord := &LogFmtRecord{}
		if err = UnmarshalLogFmtRecord(r.Context, msg, lfRecord); err == nil {
			record = lfRecord
		}
	default:
		err = ErrUnexpectedFormat
	}

	return
}
