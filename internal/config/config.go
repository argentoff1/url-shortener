package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `yaml:"env" env-default:"local"`
	// При ошибке можно удалить env-required:"true"
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// MustLoad - приставка Must нужна в функциях, которые вместо возврата ошибки паникуют
func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Переменная CONFIG_PATH не установлена")
	}

	// Проверка существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Файла конфигурации %s не существует", configPath)
	}

	var cfg Config

	// Считываем файл по указанному пути
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Невозможно прочитать конфиг: %s", err)
	}

	return &cfg
}
