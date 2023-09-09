package log

import (
	"os"

	"github.com/nats-io/nats.go"
)

type RecordType int

const (
	RecordTypeTerminal = iota
	RecordTypeLogFmt   = 1
)

type Record interface {
	Type() RecordType
	Msg() *nats.Msg
	Write(file *os.File) (n int, err error)
}
