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

	// Connect to PostgreSQL database
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create database tables
	if err := database.CreateTables(db); err != nil {
		log.Fatal(err)
	}

	// Seed initial data
	if err := database.SeedData(db); err != nil {
		log.Fatal(err)
	}

	// =========================
	// Health Check
	// =========================
	http.HandleFunc("/api/health", healthHandler)

	// =========================
	// Windows
	// =========================
	http.HandleFunc("/api/windows", handlers.GetWindows(db))

	// =========================
	// Playlist
	// GET    /api/windows/{id}/playlist
	// POST   /api/windows/{id}/playlist
	// DELETE /api/windows/{id}/playlist
	// =========================
	http.HandleFunc("/api/windows/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			handlers.GetPlaylist(db)(w, r)

		case http.MethodPost:
			handlers.AddToPlaylist(db)(w, r)

		case http.MethodDelete:
			handlers.ResetPlaylist(db)(w, r)

		default:
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// =========================
	// Media
	// GET  /api/media
	// POST /api/media
	// =========================
	http.HandleFunc("/api/media", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			handlers.GetMedia(db)(w, r)

		case http.MethodPost:
			handlers.CreateMedia(db)(w, r)

		default:
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// =========================
	// Update Media
	// PUT /api/media/{id}
	// =========================
	http.HandleFunc("/api/media/", handlers.UpdateMedia(db))

	// =========================
	// Synchronization
	// GET  /api/sync
	// POST /api/sync
	// =========================
	http.HandleFunc("/api/sync", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodPost:
			handlers.StartSync(db)(w, r)

		case http.MethodGet:
			handlers.GetSyncState(db)(w, r)

		default:
			http.Error(
				w,
				"Method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// =========================
	// Start Server
	// =========================
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	addr := "0.0.0.0:" + port

	log.Printf(
		"Media Sequencer API running on %s",
		addr,
	)

	handler := corsMiddleware(http.DefaultServeMux)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

// =========================
// Health Handler
// =========================

func healthHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status":  "ok",
		"message": "Media Sequencer API is running",
	}

	json.NewEncoder(w).Encode(response)
}

// =========================
// CORS Middleware
// =========================

func corsMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"http://localhost:5173": true,
			"http://localhost:4173": true,
		}

		// Allow deployed frontend
		frontendURL := os.Getenv("FRONTEND_URL")

		if frontendURL != "" {
			allowedOrigins[frontendURL] = true
		}

		if allowedOrigins[origin] {
			w.Header().Set(
				"Access-Control-Allow-Origin",
				origin,
			)
		}

		w.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, DELETE, OPTIONS",
		)

		w.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type",
		)

		// Handle browser preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
