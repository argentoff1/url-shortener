package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // инициализация sqlite3 драйвера
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	// Храним имя функции в которой произошла ошибка
	const op = "storage.sqlite.New"

	// Wrap error для понимания где конкретно произошла ошибка
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: В дальнейшем подключить миграции
	//
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url (
    	id INTEGER PRIMARY KEY,
    	alias TEXT NOT NULL UNIQUE,
    	url TEXT NOT NULL,
	    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}
