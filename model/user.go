package model

type User struct {
	Id         string `json:"id"`
	AccessToken string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiration int    `json:"token_expiration"`
	Session string    `json:"session"`
	FavoriteTrack []string `json:"favorite_track"`
	CreatesAt string    `json:"created_at"`
	UpdateAt string    `json:"updated_at"`
}
