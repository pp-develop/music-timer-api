package model

type Track struct {
	Uri        string   `json:"uri"`
	DurationMs int      `json:"duration_ms"`
	Isrc       string   `json:"isrc"`
	ArtistsId  []string `json:"artists_id"`
	ID         string   `json:"id,omitempty"`
}
