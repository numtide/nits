package util

import (
	"context"

	"github.com/nats-io/jwt/v2"

	"github.com/nats-io/nats.go"
)

const (
	ConnKey   = "conn"
	NKeyKey   = "nkey"
	ClaimsKey = "claims"
)

func SetClaims(ctx context.Context, claims *jwt.UserClaims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

func GetClaims(ctx context.Context) *jwt.UserClaims {
	return ctx.Value(ClaimsKey).(*jwt.UserClaims)
}

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
