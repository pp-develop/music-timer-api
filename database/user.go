package database

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func SaveRefreshToken(response model.ApiTokenResponse, id string) bool {
	_, err := db.Exec("INSERT INTO users (id, access_token, refresh_token, token_expiration) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE access_token = ?, refresh_token = ?, token_expiration=?", id, response.AccessToken, response.RefreshToken, response.ExpiresIn, response.AccessToken, response.RefreshToken, response.ExpiresIn)
	if err != nil {
		return false
	}
	return true
}

func GetUser(id string) model.User {
	var user model.User

	err := db.QueryRow("SELECT id, access_token, refresh_token FROM users WHERE id = ?", id).Scan(&user.Id, &user.AccessToken, &user.RefreshToken)
	if err != nil {
		panic(err.Error())
	}

	return user
}
