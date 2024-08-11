package database

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"golang.org/x/oauth2"
)

func SaveAccessToken(token *oauth2.Token, id string) error {
	_, err := db.Exec(`
        INSERT INTO users (id, access_token, refresh_token, token_expiration, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            token_expiration = EXCLUDED.token_expiration,
            updated_at = NOW()`,
		id, token.AccessToken, token.RefreshToken, token.Expiry.Unix())
	if err != nil {
		return err
	}
	return nil
}

func UpdateAccessToken(token *oauth2.Token, id string) error {
	_, err := db.Exec(`
        UPDATE users SET access_token = $1, updated_at = NOW()
        WHERE id = $2`,
		token.AccessToken, id)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(id string) (model.User, error) {
	var user model.User

	err := db.QueryRow(`
        SELECT id, access_token, refresh_token, token_expiration FROM users
        WHERE id = $1`, id).Scan(&user.Id, &user.AccessToken, &user.RefreshToken, &user.TokenExpiration)
	if err != nil {
		return user, err
	}

	return user, nil
}
