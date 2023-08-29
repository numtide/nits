package util

import (
	"context"

	"github.com/nats-io/nats.go"
)

const (
	ConnKey = "conn"
	NKeyKey = "nkey"
)

func SetConn(ctx context.Context, conn *nats.Conn) context.Context {
	return context.WithValue(ctx, ConnKey, conn)
}

func GetConn(ctx context.Context) *nats.Conn {
	return ctx.Value(ConnKey).(*nats.Conn)
}

func SetNKey(ctx context.Context, nkey string) context.Context {
	return context.WithValue(ctx, NKeyKey, nkey)
}

func GetNKey(ctx context.Context) string {
	return ctx.Value(NKeyKey).(string)
}
