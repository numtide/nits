package log

import (
	"fmt"

	"github.com/charmbracelet/log"
)

type NatsLogAdapter struct {
	Logger *log.Logger
}

func (n *NatsLogAdapter) Noticef(format string, v ...interface{}) {
	n.Logger.Info(fmt.Sprintf(format, v...))
}

func (n *NatsLogAdapter) Warnf(format string, v ...interface{}) {
	n.Logger.Warn(fmt.Sprintf(format, v...))
}

func (n *NatsLogAdapter) Fatalf(format string, v ...interface{}) {
	n.Logger.Fatal(fmt.Sprintf(format, v...))
}

func (n *NatsLogAdapter) Errorf(format string, v ...interface{}) {
	n.Logger.Error(fmt.Sprintf(format, v...))
}

func (n *NatsLogAdapter) Debugf(format string, v ...interface{}) {
	n.Logger.Debug(fmt.Sprintf(format, v...))
}

func (n *NatsLogAdapter) Tracef(format string, v ...interface{}) {
	n.Logger.Debug(fmt.Sprintf(format, v...))
}
