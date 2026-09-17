package database

import (
	"database/sql"
	"fmt"
)

func SeedData(db *sql.DB) error {

	// Seed media
	media := []struct {
		name      string
		mediaType string
		url       string
		duration  int
	}{
		{"M1", "image", "https://placehold.co/800x450?text=M1", 10},
		{"M2", "video", "https://www.w3schools.com/html/mov_bbb.mp4", 20},
		{"M3", "image", "https://placehold.co/800x450?text=M3", 10},
		{"M4", "video", "https://www.w3schools.com/html/movie.mp4", 20},
		{"M5", "blank", "", 10},
	}

	for _, m := range media {
		_, err := db.Exec(`
			INSERT INTO media
			(name, type, url, duration_seconds)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (name) DO NOTHING
		`, m.name, m.mediaType, m.url, m.duration)

		if err != nil {
			return fmt.Errorf("failed to seed media %s: %w", m.name, err)
		}
	}

	// Seed windows
	windows := []string{
		"Window 1",
		"Window 2",
		"Window 3",
		"Window 4",
	}

	for _, name := range windows {
		_, err := db.Exec(`
			INSERT INTO windows (name)
			VALUES ($1)
			ON CONFLICT (name) DO NOTHING
		`, name)

		if err != nil {
			return fmt.Errorf("failed to seed window %s: %w", name, err)
		}
	}

	// Get actual media IDs
	mediaIDs := make(map[string]int)

	for _, name := range []string{"M1", "M2", "M3", "M4", "M5"} {
		var id int

		err := db.QueryRow(`
			SELECT id
			FROM media
			WHERE name = $1
		`, name).Scan(&id)

		if err != nil {
			return fmt.Errorf("failed to find media %s: %w", name, err)
		}

		mediaIDs[name] = id
	}

	// Get actual window IDs
	windowIDs := make(map[string]int)

	for _, name := range windows {
		var id int

		err := db.QueryRow(`
			SELECT id
			FROM windows
			WHERE name = $1
		`, name).Scan(&id)

		if err != nil {
			return fmt.Errorf("failed to find window %s: %w", name, err)
		}

		windowIDs[name] = id
	}

	// Seed playlists
	playlists := map[string][]string{
		"Window 1": {"M1", "M2", "M3"},
		"Window 2": {"M2", "M4", "M5"},
		"Window 3": {"M1", "M3", "M4"},
		"Window 4": {"M5", "M2", "M1"},
	}

	for windowName, mediaNames := range playlists {

		windowID := windowIDs[windowName]

		for position, mediaName := range mediaNames {

			mediaID := mediaIDs[mediaName]

			_, err := db.Exec(`
				INSERT INTO playlist_items
				(window_id, media_id, position)
				VALUES ($1, $2, $3)
				ON CONFLICT (window_id, media_id, position) DO NOTHING
			`, windowID, mediaID, position)

			if err != nil {
				return fmt.Errorf(
					"failed to seed playlist for %s: %w",
					windowName,
					err,
				)
			}
		}
	}

	// Seed sync state
	_, err := db.Exec(`
		INSERT INTO sync_state
		(id, active)
		VALUES (1, 0)
		ON CONFLICT (id) DO NOTHING
	`)

	if err != nil {
		return fmt.Errorf("failed to seed sync state: %w", err)
	}

	return nil
}
