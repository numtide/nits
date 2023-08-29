package exec

import (
	"bufio"
	"bytes"

	"github.com/charmbracelet/log"
)

type Logger struct {
	Log *log.Logger

	buf []byte
}

func (l Logger) Write(b []byte) (n int, err error) {
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
