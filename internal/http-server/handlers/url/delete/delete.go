package delete

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
}

// URLDeleter - интерфейс для удаления записи из БД по alias
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias пустой")

			render.JSON(w, r, resp.Error("неверный запрос"))

			return
		}

		resultErr := deleter.DeleteURL(alias)
		if errors.Is(resultErr, storage.ErrURLNotFound) {
			log.Info("url не найден", "alias", alias)

			render.JSON(w, r, resp.Error("url не найден"))

			return
		}
		if resultErr != nil {
			log.Error("невозможно удалить url", sl.Err(resultErr))

			render.JSON(w, r, resp.Error("внутренняя ошибка"))

			return
		}

		log.Info("url удалён", "alias", alias)

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
