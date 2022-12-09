package model

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
type ApiTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// https://developer.spotify.com/documentation/web-api/reference/#/operations/create-playlist
type CreatePlaylistReqBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePlaylistResponse struct {
	ID string `json:"id"`
}
