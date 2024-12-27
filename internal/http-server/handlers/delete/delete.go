package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"url-shortener/internal/http-server/middleware/authenticator"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.0 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.0 --name=IsAdminChecker
type IsAdminChecker interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

var (
	ErrInvalidUserId = errors.New("invalid user id")
)

func New(log *slog.Logger, urlDeleter URLDeleter, isAdminChecker IsAdminChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userIdAny := r.Context().Value(authenticator.UserIdCtxKey)

		userId, ok := userIdAny.(int64)
		if !ok {
			log.Info(
				"failed to get userId from context",
				sl.Err(ErrInvalidUserId),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		isAdmin, err := isAdminChecker.IsAdmin(context.Background(), userId)

		if err != nil {
			log.Info(
				"failed to check if user isAdmin",
				sl.Err(err),
			)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal error"))

			return
		}

		if !isAdmin {
			log.Info("user isn't admin")

			render.Status(r, http.StatusForbidden)
			render.JSON(w, r, resp.Error("you are not admin"))

			return
		}

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err = urlDeleter.DeleteURL(alias)
		if err != nil {
			log.Info("failed to delete url", "alias", alias, "error", err)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "internal error")

			return
		}

		log.Info("url deleted", "alias", alias)

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, struct{}{})
	}
}
