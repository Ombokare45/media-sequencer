package models

type PlaylistItem struct {
	ID              int    `json:"id"`
	WindowID        int    `json:"window_id"`
	MediaID         int    `json:"media_id"`
	Position        int    `json:"position"`
	MediaName       string `json:"media_name"`
	MediaType       string `json:"media_type"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}
