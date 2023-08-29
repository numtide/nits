package log

import (
	"bufio"
	"bytes"

	"github.com/charmbracelet/log"
)

type BufferedLogger struct {
	Log *log.Logger

	buf []byte
}

func (l BufferedLogger) Write(b []byte) (n int, err error) {
	buf := l.buf

	buf = append(buf, b...)
	scanner := bufio.NewScanner(bytes.NewBuffer(buf))

	count := 0

	for scanner.Scan() {
		msg := scanner.Text()
		count += len(msg)
		l.Log.Info(msg)
	}

	// resize based on number of bytes read
	l.buf = buf[count:]

	return len(b), nil
}
