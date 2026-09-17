package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"media-sequencer/internal/models"
)

func GetMedia(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, name, type, url, duration_seconds
			FROM media
			ORDER BY id
		`)

		if err != nil {
			http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		mediaList := []models.Media{}

		for rows.Next() {

			var media models.Media

			err := rows.Scan(
				&media.ID,
				&media.Name,
				&media.Type,
				&media.URL,
				&media.DurationSeconds,
			)

			if err != nil {
				http.Error(w, "Failed to read media data", http.StatusInternalServerError)
				return
			}

			mediaList = append(mediaList, media)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(mediaList)
	}
}
func CreateMedia(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var media models.Media

		err := json.NewDecoder(r.Body).Decode(&media)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		result, err := db.Exec(`
			INSERT INTO media
			(name, type, url, duration_seconds)
			VALUES (?, ?, ?, ?)
		`,
			media.Name,
			media.Type,
			media.URL,
			media.DurationSeconds,
		)

		if err != nil {
			http.Error(w, "Failed to create media", http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get media ID", http.StatusInternalServerError)
			return
		}

		media.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(media)
	}
}
