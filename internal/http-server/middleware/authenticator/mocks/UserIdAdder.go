package mocks

import (
	"context"
	"net/http"
	"url-shortener/internal/http-server/middleware/authenticator"
)

func UserIdAdder(userId int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(
				r.Context(),
				authenticator.UserIdCtxKey,
				userId,
			)
			next.ServeHTTP(w, r.WithContext(ctx))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}
