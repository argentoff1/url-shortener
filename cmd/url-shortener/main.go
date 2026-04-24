package main

import (
	"fmt"
	"log/slog"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/storage/mysql"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// Файл, который запускает программу
func main() {
	// init config: cleanenv
	cfg := config.MustLoad()

	// init logger: sl
	log := setupLogger(cfg.Env)

	// Постоянный вывод наименования окружения при запуске
	// log = log.With(sl.String("env", cfg.Env))

	// Будет выведено какое окружение используется при запуске программы
	log.Info("Запуск url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug сообщения включены")

	// TODO: init storage: mysql
	storage, err := mysql.New(cfg.Dsn)
	if err != nil {
		// Собственная функция Err для slog
		log.Error("Инициализация БД провалена", sl.Err(err))
		os.Exit(1)
	}
	url, err := storage.GetURL("googl")
	if err != nil {
		log.Error("Не удалось получить URL", sl.Err(err))
	}
	fmt.Println(url)
	// TODO: init router: chi, "chi render"

	// TODO: run server
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	// Реализация разного уровня логирования для определенных окружений
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
