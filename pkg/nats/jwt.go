package nats

import (
	"io"
	"os"

	"github.com/nats-io/jwt/v2"
)

type claimsDecoder[T any] func(string) (*T, error)

func DecodeClaims[T any](str string, decoder claimsDecoder[T]) (claims *T, err error) {
	return decoder(str)
}

func DecodeUserClaims(str string) (*jwt.UserClaims, error) {
	return DecodeClaims[jwt.UserClaims](str, jwt.DecodeUserClaims)
}

func ReadClaims[T any](path string, decoder claimsDecoder[T]) (claims *T, err error) {
	var b []byte
	if b, err = os.ReadFile(path); err != nil {
		return
	}
	return DecodeClaims[T](string(b), decoder)
}

func ReadClaimsFile[T any](file *os.File, decoder claimsDecoder[T]) (claims *T, err error) {
	var b []byte
	if b, err = io.ReadAll(file); err != nil {
		return
	}
	return DecodeClaims[T](string(b), decoder)
}

func ReadUserClaims(path string) (*jwt.UserClaims, error) {
	return ReadClaims[jwt.UserClaims](path, jwt.DecodeUserClaims)
}
