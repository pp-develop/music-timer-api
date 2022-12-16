package database

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func SaveAccessToken(response model.ApiTokenResponse, id string) error {
	_, err := db.Exec("INSERT INTO users (id, access_token, refresh_token, token_expiration, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW()) ON DUPLICATE KEY UPDATE access_token = ?, refresh_token = ?, token_expiration=?, updated_at=NOW()", id, response.AccessToken, response.RefreshToken, response.ExpiresIn, response.AccessToken, response.RefreshToken, response.ExpiresIn)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(id string) (model.User, error) {
	var user model.User

	err := db.QueryRow("SELECT id, access_token, refresh_token FROM users WHERE id = ?", id).Scan(&user.Id, &user.AccessToken, &user.RefreshToken)
	if err != nil {
		return user, err
	}

	return user, nil
}
