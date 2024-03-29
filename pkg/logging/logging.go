package logging

import (
	"os"

	"github.com/nats-io/nats.go"
)

type RecordType int

const (
	RecordTerm   = iota
	RecordLogFmt = 1
)

type Record interface {
	Type() RecordType
	Msg() *nats.Msg
	Write(file *os.File) (n int, err error)
}
