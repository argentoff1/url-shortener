package redirect

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

// URLGetter - интерфейс для получения url по alias.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

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

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url не найден", "alias", alias)

			render.JSON(w, r, resp.Error("url не найден"))

			return
		}
		if err != nil {
			log.Error("невозможно получить url", sl.Err(err))

			render.JSON(w, r, resp.Error("внутренняя ошибка"))

			return
		}

		log.Info("url получен", slog.String("url", resURL))

		// redirect на найденный url
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
