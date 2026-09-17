package main

import (
	"encoding/json"
	"log"
	"media-sequencer/internal/database"
	"media-sequencer/internal/handlers"
	"net/http"
	"os"
)

func main() {

	// Connect to SQLite database
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database tables
	if err := database.CreateTables(db); err != nil {
		log.Fatal(err)
	}

	if err := database.SeedData(db); err != nil {
		log.Fatal(err)
	}

	// Health check endpoint
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/windows", handlers.GetWindows(db))

	http.HandleFunc("/api/windows/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetPlaylist(db)(w, r)
			return
		}

		if r.Method == http.MethodPost {
			handlers.AddToPlaylist(db)(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			handlers.ResetPlaylist(db)(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	http.HandleFunc("/api/media", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			handlers.GetMedia(db)(w, r)
			return
		}

		if r.Method == http.MethodPost {
			handlers.CreateMedia(db)(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {
			handlers.StartSync(db)(w, r)
			return
		}

		if r.Method == http.MethodGet {
			handlers.GetSyncState(db)(w, r)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
	log.Println("Media Sequencer API running on :8080")

	// Start server
	handler := corsMiddleware(http.DefaultServeMux)

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status":  "ok",
		"message": "Media Sequencer API is running",
	}

	json.NewEncoder(w).Encode(response)
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		allowedOrigin := os.Getenv("FRONTEND_URL")

		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:5173"
		}

		if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})

}
