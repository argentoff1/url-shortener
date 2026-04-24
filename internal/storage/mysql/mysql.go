package mysql

import (
	"database/sql"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql" // инициализация mysql драйвера
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	// Храним имя функции в которой произошла ошибка
	const op = "storage.mysql.New"

	db, openErr := sql.Open("mysql", dsn)
	if openErr != nil {
		return nil, fmt.Errorf("%s: %w", op, openErr)
	}

	// TODO: В дальнейшем подключить миграции
	_, execErr := db.Exec(`
	CREATE TABLE IF NOT EXISTS url (
    	id INTEGER PRIMARY KEY AUTO_INCREMENT,
    	alias varchar(255) NOT NULL UNIQUE,
    	url TEXT NOT NULL,
	    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);
	`)
	if execErr != nil {
		return nil, fmt.Errorf("%s: %w", op, execErr)
	}

	// Проверка наличия индекса
	var indexExists bool
	row := db.QueryRow(`
		SELECT COUNT(1) > 0 FROM information_schema.statistics WHERE 
        table_schema = DATABASE() AND table_name = 'url' AND index_name = 'idx_alias'
	`)
	err := row.Scan(&indexExists)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if !indexExists {
		// Создаем индекс, если его нет
		_, err = db.Exec(`CREATE INDEX idx_alias ON url(alias);`)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return &Storage{db: db}, nil
}

// SaveURL - создаем запись в БД, возвращаем ID добавленного URL
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.mysql.SaveURL"

	stmt, err := s.db.Prepare(`INSERT INTO url (url, alias) VALUES (?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// TODO: переписать данную реализацию
		// Код 1062 -Нарушение ограничения уникальности
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: не удалось получить последний добавленный ID: %w", op, err)
	}

	return id, nil
}
