package database

import (
	"database/sql"
	"log/slog"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func SaveAccessToken(db *sql.DB, token *oauth2.Token, user *spotify.PrivateUser) error {
	// トークンを暗号化
	encryptedAccessToken, err := utils.EncryptToken(token.AccessToken)
	if err != nil {
		slog.Error("failed to encrypt access token", slog.Any("error", err))
		return err
	}

	encryptedRefreshToken, err := utils.EncryptToken(token.RefreshToken)
	if err != nil {
		slog.Error("failed to encrypt refresh token", slog.Any("error", err))
		return err
	}

	_, err = db.Exec(`
        INSERT INTO spotify_users (id, country, access_token, refresh_token, token_expiration, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
			country = EXCLUDED.country,
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            token_expiration = EXCLUDED.token_expiration,
            updated_at = NOW()`,
		user.ID, user.Country, encryptedAccessToken, encryptedRefreshToken, token.Expiry.Unix())
	if err != nil {
		return err
	}
	return nil
}

func UpdateAccessToken(db *sql.DB, token *oauth2.Token, id string) error {
	// アクセストークンを暗号化
	encryptedAccessToken, err := utils.EncryptToken(token.AccessToken)
	if err != nil {
		slog.Error("failed to encrypt access token", slog.Any("error", err))
		return err
	}

	_, err = db.Exec(`
        UPDATE spotify_users SET access_token = $1, updated_at = NOW()
        WHERE id = $2`,
		encryptedAccessToken, id)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(db *sql.DB, id string) (model.User, error) {
	var user model.User
	var encryptedAccessToken, encryptedRefreshToken string

	err := db.QueryRow(`
        SELECT id, country, access_token, refresh_token, token_expiration, updated_at FROM spotify_users
        WHERE id = $1`, id).Scan(&user.Id, &user.Country, &encryptedAccessToken, &encryptedRefreshToken, &user.TokenExpiration, &user.UpdateAt)
	if err != nil {
		return user, err
	}

	// トークンを復号化
	user.AccessToken, err = utils.DecryptToken(encryptedAccessToken)
	if err != nil {
		slog.Error("failed to decrypt access token", slog.Any("error", err))
		return user, err
	}

	user.RefreshToken, err = utils.DecryptToken(encryptedRefreshToken)
	if err != nil {
		slog.Error("failed to decrypt refresh token", slog.Any("error", err))
		return user, err
	}

	return user, nil
}

func IncrementPlaylistCount(db *sql.DB, id string) error {
	_, err := db.Exec(`
        UPDATE spotify_users SET playlist_count = playlist_count + 1, updated_at = NOW()
        WHERE id = $1`,
		id)
	if err != nil {
		return err
	}
	return nil
}
