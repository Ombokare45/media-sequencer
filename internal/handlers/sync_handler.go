package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

func StartSync(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var request struct {
			MediaID         int `json:"media_id"`
			DurationSeconds int `json:"duration_seconds"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if request.MediaID <= 0 {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
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

		// Verify that the media exists.
		var mediaExists int

		err := db.QueryRow(`
			SELECT id
			FROM media
			WHERE id = $1
		`, request.MediaID).Scan(&mediaExists)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Media not found", http.StatusNotFound)
				return
			}

			http.Error(
				w,
				"Failed to verify media",
				http.StatusInternalServerError,
			)
			return
		}

		startedAt := time.Now().UTC()

		_, err = db.Exec(`
			UPDATE sync_state
			SET
				media_id = $1,
				started_at = $2,
				duration_seconds = $3,
				active = 1
			WHERE id = 1
		`,
			request.MediaID,
			startedAt,
			request.DurationSeconds,
		)

		if err != nil {
			http.Error(
				w,
				"Failed to start sync",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":          "Sync started successfully",
			"media_id":         request.MediaID,
			"started_at":       startedAt,
			"duration_seconds": request.DurationSeconds,
			"active":           true,
		})
	}
}

func GetSyncState(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var (
			mediaID         sql.NullInt64
			mediaName       sql.NullString
			mediaType       sql.NullString
			mediaURL        sql.NullString
			startedAt       sql.NullTime
			durationSeconds int
			active          int
		)

		err := db.QueryRow(`
			SELECT
				s.media_id,
				m.name,
				m.type,
				m.url,
				s.started_at,
				s.duration_seconds,
				s.active
			FROM sync_state s
			LEFT JOIN media m ON s.media_id = m.id
			WHERE s.id = 1
		`).Scan(
			&mediaID,
			&mediaName,
			&mediaType,
			&mediaURL,
			&startedAt,
			&durationSeconds,
			&active,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Sync state not found", http.StatusNotFound)
				return
			}

			http.Error(
				w,
				"Failed to fetch sync state",
				http.StatusInternalServerError,
			)
			return
		}

		// Automatically expire the sync when its duration ends.
		if active == 1 && startedAt.Valid {

			elapsed := time.Since(startedAt.Time).Seconds()

			if elapsed >= float64(durationSeconds) {

				_, err := db.Exec(`
					UPDATE sync_state
					SET active = 0
					WHERE id = 1
				`)

				if err != nil {
					http.Error(
						w,
						"Failed to update sync state",
						http.StatusInternalServerError,
					)
					return
				}

				active = 0
			}
		}

		response := map[string]interface{}{
			"media_id":         nil,
			"media_name":       nil,
			"type":             nil,
			"url":              nil,
			"started_at":       nil,
			"duration_seconds": durationSeconds,
			"active":           active == 1,
		}

		if mediaID.Valid {
			response["media_id"] = mediaID.Int64
		}

		if mediaName.Valid {
			response["media_name"] = mediaName.String
		}

		if mediaType.Valid {
			response["type"] = mediaType.String
		}

		if mediaURL.Valid {
			response["url"] = mediaURL.String
		}

		if startedAt.Valid {
			response["started_at"] = startedAt.Time
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response)
	}
}
