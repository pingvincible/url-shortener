package authenticator

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/jwtauth/v5"
)

var (
	ErrEmptyClaims = errors.New("empty claims")
	ErrEmptyToken  = errors.New("empty token")
)

type contextKey struct {
	name string
}

var (
	UserIdCtxKey = &contextKey{"UserId"}
)

func Authenticator(log *slog.Logger, ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, claims, err := jwtauth.FromContext(r.Context())

			if err != nil {
				log.Info("failed to extract token", sl.Err(err))
				responseUnauthorized(w, r)
				return
			}

			if token == nil {
				log.Info("token is nil", sl.Err(ErrEmptyToken))
				responseUnauthorized(w, r)
				return
			}

			if claims == nil {
				log.Info("claims are nil", sl.Err(ErrEmptyClaims))
				responseUnauthorized(w, r)
				return
			}

			fmt.Printf("%+v", claims)
			fmt.Printf("%+v", r.Context())

			ctx := r.Context()
			userId := int64(claims["uid"].(float64))
			ctx = context.WithValue(
				ctx,
				UserIdCtxKey,
				userId,
			)

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(hfn)
	}
}

func responseUnauthorized(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, resp.Response{
		Status: resp.StatusError,
		Error:  "Unauthorized",
	})
}
