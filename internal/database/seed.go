package database

import "database/sql"

func SeedData(db *sql.DB) error {

	// Add media
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
			INSERT OR IGNORE INTO media
			(name, type, url, duration_seconds)
			VALUES (?, ?, ?, ?)
		`, m.name, m.mediaType, m.url, m.duration)

		if err != nil {
			return err
		}
	}

	// Add windows
	windows := []string{
		"Window 1",
		"Window 2",
		"Window 3",
		"Window 4",
	}

	for _, name := range windows {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO windows (name)
			VALUES (?)
		`, name)

		if err != nil {
			return err
		}
	}

	// Add playlists
	playlists := map[int][]int{
		1: {1, 2, 3},
		2: {2, 4, 5},
		3: {1, 3, 4},
		4: {5, 2, 1},
	}

	for windowID, mediaIDs := range playlists {

		for position, mediaID := range mediaIDs {

			_, err := db.Exec(`
				INSERT OR IGNORE INTO playlist_items
				(window_id, media_id, position)
				VALUES (?, ?, ?)
			`, windowID, mediaID, position)

			if err != nil {
				return err
			}
		}
	}

	// Initialize sync state
	_, err := db.Exec(`
		INSERT OR IGNORE INTO sync_state
		(id, active)
		VALUES (1, 0)
	`)

	if err != nil {
		return err
	}

	return nil
}
