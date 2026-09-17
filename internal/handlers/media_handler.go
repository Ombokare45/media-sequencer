package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"media-sequencer/internal/models"
)

func GetMedia(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		rows, err := db.Query(`
			SELECT id, name, type, url, duration_seconds
			FROM media
			ORDER BY id
		`)

		if err != nil {
			http.Error(
				w,
				"Failed to fetch media",
				http.StatusInternalServerError,
			)
			return
		}

		defer rows.Close()

		mediaList := []models.Media{}

		for rows.Next() {

			var media models.Media

			if err := rows.Scan(
				&media.ID,
				&media.Name,
				&media.Type,
				&media.URL,
				&media.DurationSeconds,
			); err != nil {
				http.Error(
					w,
					"Failed to read media",
					http.StatusInternalServerError,
				)
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

		var request struct {
			Name            string `json:"name"`
			Type            string `json:"type"`
			URL             string `json:"url"`
			DurationSeconds int    `json:"duration_seconds"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(
				w,
				"Invalid request body",
				http.StatusBadRequest,
			)
			return
		}

		if request.Name == "" {
			http.Error(w, "Media name is required", http.StatusBadRequest)
			return
		}

		if request.Type == "" {
			http.Error(w, "Media type is required", http.StatusBadRequest)
			return
		}

		if request.DurationSeconds <= 0 {
			http.Error(
				w,
				"Duration must be greater than 0",
				http.StatusBadRequest,
			)
			return
		}

		var media models.Media

		err := db.QueryRow(`
			INSERT INTO media
				(name, type, url, duration_seconds)
			VALUES ($1, $2, $3, $4)
			RETURNING id, name, type, url, duration_seconds
		`,
			request.Name,
			request.Type,
			request.URL,
			request.DurationSeconds,
		).Scan(
			&media.ID,
			&media.Name,
			&media.Type,
			&media.URL,
			&media.DurationSeconds,
		)

		if err != nil {
			http.Error(
				w,
				"Failed to create media",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(media)
	}
}
