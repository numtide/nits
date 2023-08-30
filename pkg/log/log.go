package log

import (
	"bufio"
	"github.com/charmbracelet/log"
	"io"
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
