package model

type Track struct {
	Uri         string   `json:"uri"`
	DurationMs  int      `json:"duration_ms"`
	ArtistsId   []string `json:"artists_id"`
	ArtistsName []string `json:"artists_name"`
}
