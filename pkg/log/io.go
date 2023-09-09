package log

import (
	"context"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	nnats "github.com/numtide/nits/pkg/nats"
)

const (
	ErrUnexpectedFormat = errors.ConstError("unexpected format")
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
	case HeaderFormatTerminal:
		tRecord := &TerminalRecord{}
		if err = UnmarshalTerminalRecord(msg, tRecord); err == nil {
			record = tRecord
		}
	case HeaderFormatLogFmt:
		lfRecord := &LogFmtRecord{}
		if err = UnmarshalLogFmtRecord(msg, lfRecord); err == nil {
			record = lfRecord
		}
	default:
		err = ErrUnexpectedFormat
	}

	return
}
