package mysql

import (
	"database/sql"
	"fmt"

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

	// Wrap error для понимания где конкретно произошла ошибка
	/*db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}*/

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
