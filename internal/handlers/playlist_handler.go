package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"media-sequencer/internal/models"
)

func GetPlaylist(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		windowID, err := getWindowID(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
			SELECT
				pi.id,
				pi.window_id,
				pi.media_id,
				pi.position,
				m.name,
				m.type,
				m.url,
				m.duration_seconds
			FROM playlist_items pi
			JOIN media m ON pi.media_id = m.id
			WHERE pi.window_id = $1
			ORDER BY pi.position
		`, windowID)

		if err != nil {
			http.Error(
				w,
				"Failed to fetch playlist",
				http.StatusInternalServerError,
			)
			return
		}

		defer rows.Close()

		playlist := []models.PlaylistItem{}

		for rows.Next() {

			var item models.PlaylistItem

			if err := rows.Scan(
				&item.ID,
				&item.WindowID,
				&item.MediaID,
				&item.Position,
				&item.MediaName,
				&item.MediaType,
				&item.URL,
				&item.DurationSeconds,
			); err != nil {
				http.Error(
					w,
					"Failed to read playlist",
					http.StatusInternalServerError,
				)
				return
			}

			playlist = append(playlist, item)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(playlist)
	}
}

func AddToPlaylist(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		windowID, err := getWindowID(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		var request struct {
			MediaID int `json:"media_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if request.MediaID <= 0 {
			http.Error(w, "Invalid media ID", http.StatusBadRequest)
			return
		}

		var position int

		err = db.QueryRow(`
			SELECT COALESCE(MAX(position), -1) + 1
			FROM playlist_items
			WHERE window_id = $1
		`, windowID).Scan(&position)

		if err != nil {
			http.Error(
				w,
				"Failed to determine playlist position",
				http.StatusInternalServerError,
			)
			return
		}

		var item models.PlaylistItem

		err = db.QueryRow(`
			INSERT INTO playlist_items
				(window_id, media_id, position)
			VALUES ($1, $2, $3)
			RETURNING id, window_id, media_id, position
		`, windowID, request.MediaID, position).Scan(
			&item.ID,
			&item.WindowID,
			&item.MediaID,
			&item.Position,
		)

		if err != nil {
			http.Error(
				w,
				"Failed to add media to playlist",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(item)
	}
}

func ResetPlaylist(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		windowID, err := getWindowID(r.URL.Path)
		if err != nil {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			DELETE FROM playlist_items
			WHERE window_id = $1
		`, windowID)

		if err != nil {
			http.Error(
				w,
				"Failed to reset playlist",
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Playlist reset successfully",
		})
	}
}

func getWindowID(path string) (int, error) {

	parts := strings.Split(
		strings.Trim(path, "/"),
		"/",
	)

	// Expected:
	// /api/windows/{id}/playlist
	if len(parts) < 4 {
		return 0, strconv.ErrSyntax
	}

	return strconv.Atoi(parts[2])
}
