package main

import (
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/update"
	"url-shortener/internal/http-server/handlers/url/delete"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/middleware/mwLogger"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/mysql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// init config: cleanenv
	cfg := config.MustLoad()

	// init logger: sl
	log := setupLogger(cfg.Env)

	// Постоянный вывод наименования окружения при запуске
	// log = log.With(sl.String("env", cfg.Env))

	// Будет выведено какое окружение используется при запуске программы
	log.Info(
		"Запуск url-shortener",
		slog.String("env", cfg.Env),
		slog.String("version", "0.1"),
	)
	log.Debug("debug сообщения включены")

	// init storage: mysql
	storage, err := mysql.New(cfg.Dsn)
	if err != nil {
		// Собственная функция Err для slog
		log.Error("Инициализация БД провалена", sl.Err(err))
		os.Exit(1)
	}

	// init router: chi
	router := chi.NewRouter()

	// Middleware
	// Добавление к каждому запросу request id. Для трейсинга
	router.Use(middleware.RequestID)
	// Логирование всех входящих запросов
	router.Use(mwLogger.New(log))
	// Если срабатывает panic, восстанавливаем приложение
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Создаем внутри роутера еще один, для подключения авторизации
	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))

		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
		r.Put("/update", update.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))

	// run server
	log.Info("Запуск сервера", slog.String("address", cfg.Address))

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("Не удалось запустить сервер")
	}

	log.Error("Сервер остановлен")

	defer storage.Close()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	// Реализация разного уровня логирования для определенных окружений
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
