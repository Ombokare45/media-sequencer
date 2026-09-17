package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func Connect() (*sql.DB, error) {
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := os.Getenv("DB_PATH")

	if dbPath == "" {
		dbPath = "data/sequencer.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}
