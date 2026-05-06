package update

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	OldAlias string `json:"oldAlias" validate:"required"` // Validate - дает информацию валидатору
	NewAlias string `json:"newAlias" validate:"required"`
}

/*
Response

	Status: Error / Ok
	Alias: Если в запросе alias - nil, тогда он будет генерироваться из случайных символов
*/
type Response struct {
	resp.Response
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UrlUpdater
type UrlUpdater interface {
	UpdateURL(newAlias string, oldAlias string) error
}

/*
New - конструктор для хэндлера
Во время подключения хэндлера к роутеру будет вызываться данная функция
*/
func New(log *slog.Logger, urlUpdater UrlUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		// Распарсим запрос
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("Не удалось расшифровать тело запроса", sl.Err(err))

			// Возвращаем JSON с ответом для клиента
			render.JSON(w, r, resp.Error("Не удалось расшифровать запрос"))

			return
		}

		log.Info("Тело запроса расшифровано", slog.Any("request", req))

		// Валидация запроса
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("Недопустимый запрос", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		// TODO: реализовать проверку alias на уже существующий
		oldAlias := req.OldAlias
		newAlias := req.NewAlias

		err = urlUpdater.UpdateURL(newAlias, oldAlias)
		if errors.Is(err, storage.ErrAliasNotFound) {
			log.Info("alias не найден", slog.String("url", req.OldAlias))

			render.JSON(w, r, resp.Error("alias не найден"))

			return
		}
		if err != nil {
			log.Error("Не удалось обновить alias", sl.Err(err))

			render.JSON(w, r, resp.Error("Не удалось обновить alias"))

			return
		}

		log.Info("alias обновлен", slog.String("newAlias", newAlias))

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}
