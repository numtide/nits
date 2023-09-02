package log

import (
	"bufio"
	"bytes"
	"io"

	"github.com/go-logfmt/logfmt"
	"github.com/nats-io/nats.go"

	"github.com/charmbracelet/log"
)

type Writer struct {
	Log    *log.Logger
	reader *io.PipeReader
	writer *io.PipeWriter
}

func (w *Writer) Close() (err error) {
	if w.writer != nil {
		err = w.writer.Close()
	}
	return
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.reader == nil {
		w.reader, w.writer = io.Pipe()
		go w.process()
	}
	return w.writer.Write(p)
}

func (w *Writer) process() {
	scanner := bufio.NewScanner(w.reader)
	for scanner.Scan() {
		line := scanner.Text()
		w.Log.Info(line)
	}
}

func MsgToLog(logger *log.Logger, msg *nats.Msg) {
	if msg == nil {
		return
	}
	dec := logfmt.NewDecoder(bytes.NewReader(msg.Data))
	for dec.ScanRecord() {

		var message, lvl string
		var kvs []interface{}

		for dec.ScanKeyval() {
			key := string(dec.Key())
			value := string(dec.Value())
			switch key {
			case "msg":
				message = value
			case "lvl":
				lvl = value
			default:
				kvs = append(kvs, key)
				kvs = append(kvs, string(dec.Value()))
			}
		}

		switch lvl {
		case "warn":
			logger.Warn(message, kvs...)
		case "info":
			logger.Info(message, kvs...)
		case "debug":
			logger.Debug(message, kvs...)
		case "error", "fatal":
			logger.Error(message, kvs...)
			logger.Error(message, kvs...)
		default:
			logger.Info(message, kvs...)
		}
	}
}
