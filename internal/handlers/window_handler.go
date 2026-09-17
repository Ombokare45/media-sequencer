package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"media-sequencer/internal/models"
)

func GetWindows(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(`
			SELECT id, name
			FROM windows
			ORDER BY id
		`)

		if err != nil {
			http.Error(w, "Failed to fetch windows", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var windows []models.Window

		for rows.Next() {

			var window models.Window

			err := rows.Scan(
				&window.ID,
				&window.Name,
			)

			if err != nil {
				http.Error(w, "Failed to read window data", http.StatusInternalServerError)
				return
			}

			windows = append(windows, window)
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(windows)
	}
}
