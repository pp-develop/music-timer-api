package model

type ApiTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type CreatePlaylistReqBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePlaylistResponse struct {
	ID string `json:"id"`
}
