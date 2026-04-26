package save

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"` // validate - дает информацию валидатору
	Alias string `json:"alias,omitempty"`
}

/*
Response

	Status: Error / Ok
	Alias: Если в запросе alias - nil, тогда он будет генерироваться из случайных символов
*/
type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// TODO: перенести в config
const aliasLength = 7

type UrlSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

/*
New - конструктор для хэндлера
Во время подключения хэндлера к роутеру будет вызываться данная функция
*/
func New(log *slog.Logger, urlSaver UrlSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

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
		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("URL уже существует", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("URL уже существует"))

			return
		}
		if err != nil {
			log.Error("Не удалось сохранить URL", sl.Err(err))

			render.JSON(w, r, resp.Error("Не удалось сохранить URL"))

			return
		}

		log.Info("URL сохранен", slog.Int64("id", id))

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
