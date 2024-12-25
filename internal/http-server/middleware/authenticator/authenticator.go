package authenticator

import (
	"github.com/go-chi/render"
	"net/http"
	resp "url-shortener/internal/lib/api/response"

	"github.com/go-chi/jwtauth/v5"
)

func Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				responseUnauthorized(w, r)
				return
			}

			if token == nil {
				responseUnauthorized(w, r)
				return
			}

			// Token is authenticated, pass it through
			next.ServeHTTP(w, r)
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
