package log

import (
	"errors"
	"fmt"
	"github.com/nats-io/nats.go"
	"os"
	"time"
)

type RecordType int

const (
	RecordTypeTerminal = iota
	RecordTypeLogFmt   = 1
)

type EOS struct {
	Subject string
}

func (e EOS) Error() string {
	return fmt.Sprintf("end of stream: %s", e.Subject)
}

func IsEOS(err error) bool {
	var EOS EOS
	return errors.As(err, &EOS)
}

type Record interface {
	Type() RecordType
	Msg() *nats.Msg
	Write(file *os.File) (n int, err error)
}

type RecordReader struct {
	Sub     *nats.Subscription
	Timeout time.Duration
}

func (r *RecordReader) Read() (record Record, err error) {
	if r.Timeout == 0 {
		r.Timeout = DefaultReadTimeout
	}
	var msg *nats.Msg
	if msg, err = r.Sub.NextMsg(r.Timeout); err != nil {
		return
	}

	if msg.Header.Get(HeaderEOF) == HeaderEOFValue {
		return nil, EOS{Subject: msg.Subject}
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
