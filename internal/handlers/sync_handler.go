package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

func StartSync(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			MediaID         int `json:"media_id"`
			DurationSeconds int `json:"duration_seconds"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if request.MediaID <= 0 {
			http.Error(w, "Invalid media_id", http.StatusBadRequest)
			return
		}

		if request.DurationSeconds <= 0 {
			http.Error(w, "Invalid duration_seconds", http.StatusBadRequest)
			return
		}

		// Check that media exists
		var mediaName string

		err = db.QueryRow(`
			SELECT name
			FROM media
			WHERE id = ?
		`, request.MediaID).Scan(&mediaName)

		if err == sql.ErrNoRows {
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, "Failed to check media", http.StatusInternalServerError)
			return
		}

		// Use UTC so every display window has the same reference time.
		startedAt := time.Now().UTC()

		// Store the active sync state.
		_, err = db.Exec(`
			UPDATE sync_state
			SET media_id = ?,
				started_at = ?,
				duration_seconds = ?,
				active = 1
			WHERE id = 1
		`,
			request.MediaID,
			startedAt,
			request.DurationSeconds,
		)

		if err != nil {
			http.Error(w, "Failed to start sync", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"status":           "sync_started",
			"media_id":         request.MediaID,
			"media_name":       mediaName,
			"started_at":       startedAt,
			"duration_seconds": request.DurationSeconds,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(response)
	}
}
func GetSyncState(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var (
			mediaID         sql.NullInt64
			startedAt       sql.NullString
			durationSeconds int
			active          int
			mediaName       sql.NullString
			mediaType       sql.NullString
			mediaURL        sql.NullString
		)

		err := db.QueryRow(`
			SELECT
				s.media_id,
				s.started_at,
				s.duration_seconds,
				s.active,
				m.name,
				m.type,
				m.url
			FROM sync_state s
			LEFT JOIN media m ON s.media_id = m.id
			WHERE s.id = 1
		`).Scan(
			&mediaID,
			&startedAt,
			&durationSeconds,
			&active,
			&mediaName,
			&mediaType,
			&mediaURL,
		)

		if err != nil {
			http.Error(w, "Failed to fetch sync state", http.StatusInternalServerError)
			return
		}

		// Check whether the sync duration has expired.
		if active == 1 && startedAt.Valid {

			startTime, err := time.Parse(time.RFC3339Nano, startedAt.String)
			if err == nil {

				elapsed := time.Since(startTime).Seconds()

				if elapsed >= float64(durationSeconds) {

					// Mark sync as inactive.
					_, err = db.Exec(`
						UPDATE sync_state
						SET active = 0
						WHERE id = 1
					`)

					if err == nil {
						active = 0
					}
				}
			}
		}

		response := map[string]interface{}{
			"active":           active == 1,
			"media_id":         mediaID.Int64,
			"media_name":       mediaName.String,
			"media_type":       mediaType.String,
			"url":              mediaURL.String,
			"started_at":       startedAt.String,
			"duration_seconds": durationSeconds,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
