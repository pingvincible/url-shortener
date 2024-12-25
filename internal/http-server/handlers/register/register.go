package register

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
}

//go:generate go run github.com/vektra/mockery/v2@v2.50.0 --name=UserRegisterer
type UserRegisterer interface {
	Register(ctx context.Context, email, password string) (int64, error)
}

func New(log *slog.Logger, userRegisterer UserRegisterer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.register.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		userId, err := userRegisterer.Register(context.Background(), req.Email, req.Password)
		if err != nil {
			log.Error("failed to register user", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, fmt.Errorf("internal server error"))

			return
		}

		log.Info("user registered", slog.Int64("user_id", userId))

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, struct{}{})
	}
}
