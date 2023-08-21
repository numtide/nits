package cache

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nats-server/v2/server"
)

const (
	AccountKey = "account"
	jwtKey     = "jwt"
)

func ClientInfo(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var err error
		var account string

		if header := r.Header.Get(server.ClientInfoHdr); header != "" {
			var clientInfo server.ClientInfo
			if err = json.Unmarshal([]byte(header), &clientInfo); err != nil {
				// todo better error feedback / logging
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("malformed header: " + server.ClientInfoHdr))
				return
			}

			account = clientInfo.Account
		}

		if _, jwtStr, ok := r.BasicAuth(); ok {

			var claims *jwt.UserClaims
			if claims, err = jwt.DecodeUserClaims(jwtStr); err != nil {
				// todo better error feedback / logging
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("malformed jwt"))
				return
			}

			var vr jwt.ValidationResults
			claims.Validate(&vr)
			if !vr.IsEmpty() {
				// todo better error feedback / logging
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// in the event of a signing key issuer account will be set, otherwise just issuer
			if account = claims.IssuerAccount; account == "" {
				account = claims.Issuer
			}
		}

		if account == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx = context.WithValue(ctx, AccountKey, account)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func GetAccount(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if account, ok := ctx.Value(AccountKey).(string); ok {
		return account
	}
	return ""
}
