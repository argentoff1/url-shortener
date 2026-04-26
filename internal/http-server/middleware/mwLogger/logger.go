package mwLogger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func New(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Создание копии логгера для middleware
		log = log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("mwLogger middleware enabled")

		// Внутренняя часть хэндлера, которая будет выполняться при каждом запросе
		fn := func(w http.ResponseWriter, r *http.Request) {
			// Данная часть выполняется ДО обработки запроса
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				// Каждому запросу присваивается request id, с которым в дальнейшем можно работать
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			// Будет выполнен после всех middleware и самого запроса
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
