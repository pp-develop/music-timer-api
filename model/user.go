package model

type User struct {
	Id              string `json:"id"`
	Country         string `json:"country"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	TokenExpiration int    `json:"token_expiration"`
	Session         string `json:"session"`
	CreatesAt       string `json:"created_at"`
	UpdateAt        string `json:"updated_at"`
}
