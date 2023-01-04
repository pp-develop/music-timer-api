package model

// https://developer.spotify.com/documentation/web-api/reference/#/operations/create-playlist
type CreatePlaylistReqBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
type CreatePlaylistResponse struct {
	ID string `json:"id"`
}
