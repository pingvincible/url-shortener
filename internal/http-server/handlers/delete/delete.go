package delete

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50.0 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
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
