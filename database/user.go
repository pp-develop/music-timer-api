package database

import(
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func SaveRefreshToken(response model.ApiTokenResponse) bool {
	_, err := db.Exec("INSERT INTO users (access_token, refresh_token, token_expiration) VALUES (?, ?, ?)", response.AccessToken, response.RefreshToken, response.ExpiresIn)
	if err != nil {
		return false
	}
	return true
}