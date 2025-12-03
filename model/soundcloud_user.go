package model

type SoundCloudUser struct {
	Id              string `json:"id"`
	Username        string `json:"username"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	TokenExpiration int    `json:"token_expiration"`
	Session         string `json:"session"`
}
