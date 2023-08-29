package util

import (
	"context"

	"github.com/charmbracelet/log"

	"github.com/nats-io/nats.go"
)

const (
	ConnKey = "conn"
	LogKey  = "log"
	NKeyKey = "nkey"
)

func SetConn(ctx context.Context, conn *nats.Conn) context.Context {
	return context.WithValue(ctx, ConnKey, conn)
}

func GetConn(ctx context.Context) *nats.Conn {
	return ctx.Value(ConnKey).(*nats.Conn)
}

func SetLog(ctx context.Context, log *log.Logger) context.Context {
	return context.WithValue(ctx, LogKey, log)
}

func GetLog(ctx context.Context) *log.Logger {
	return ctx.Value(LogKey).(*log.Logger)
}

func SetNKey(ctx context.Context, nkey string) context.Context {
	return context.WithValue(ctx, NKeyKey, nkey)
}

func GetNKey(ctx context.Context) string {
	return ctx.Value(NKeyKey).(string)
}
