package database

import (
	"database/sql"
	"fmt"
)

func CreateTables(db *sql.DB) error {

	queries := []string{

		`CREATE TABLE IF NOT EXISTS windows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			type TEXT NOT NULL,
			url TEXT,
			duration_seconds INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS playlist_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			window_id INTEGER NOT NULL,
			media_id INTEGER NOT NULL,
			position INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (window_id) REFERENCES windows(id),
			FOREIGN KEY (media_id) REFERENCES media(id),

			UNIQUE(window_id, media_id, position)
		)`,

		`CREATE TABLE IF NOT EXISTS sync_state (
			id INTEGER PRIMARY KEY,
			media_id INTEGER,
			started_at DATETIME,
			duration_seconds INTEGER DEFAULT 0,
			active INTEGER DEFAULT 0,

			FOREIGN KEY (media_id) REFERENCES media(id)
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}
