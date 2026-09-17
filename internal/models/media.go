package models

type Media struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	DurationSeconds int    `json:"duration_seconds"`
}
