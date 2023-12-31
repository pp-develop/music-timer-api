package model

type Track struct {
	Uri         string   `json:"uri"`
	DurationMs  int      `json:"duration_ms"`
	ArtistsName []string `json:"artists_name"`
}
