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

		// Get window ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) < 4 {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		windowID, err := strconv.Atoi(parts[2])
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
			WHERE pi.window_id = ?
			ORDER BY pi.position
		`, windowID)

		if err != nil {
			http.Error(w, "Failed to fetch playlist", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		playlist := []models.PlaylistItem{}

		for rows.Next() {

			var item models.PlaylistItem

			err := rows.Scan(
				&item.ID,
				&item.WindowID,
				&item.MediaID,
				&item.Position,
				&item.MediaName,
				&item.MediaType,
				&item.URL,
				&item.DurationSeconds,
			)

			if err != nil {
				http.Error(w, "Failed to read playlist data", http.StatusInternalServerError)
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

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get window ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) < 4 {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		windowID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		// Request body
		var request struct {
			MediaID int `json:"media_id"`
		}

		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if request.MediaID <= 0 {
			http.Error(w, "Invalid media_id", http.StatusBadRequest)
			return
		}

		// Find the next playlist position
		var position int

		err = db.QueryRow(`
			SELECT COALESCE(MAX(position), -1) + 1
			FROM playlist_items
			WHERE window_id = ?
		`, windowID).Scan(&position)

		if err != nil {
			http.Error(w, "Failed to determine playlist position", http.StatusInternalServerError)
			return
		}

		// Add media to playlist
		result, err := db.Exec(`
			INSERT INTO playlist_items
			(window_id, media_id, position)
			VALUES (?, ?, ?)
		`, windowID, request.MediaID, position)

		if err != nil {
			http.Error(w, "Failed to add media to playlist", http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, "Failed to get playlist item ID", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":        id,
			"window_id": windowID,
			"media_id":  request.MediaID,
			"position":  position,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(response)
	}
}
func ResetPlaylist(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) < 4 {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		windowID, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(w, "Invalid window ID", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(
			"DELETE FROM playlist_items WHERE window_id = ?",
			windowID,
		)

		if err != nil {
			http.Error(w, "Failed to reset playlist", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "success",
			"message":   "Playlist reset successfully",
			"window_id": windowID,
		})
	}
}
