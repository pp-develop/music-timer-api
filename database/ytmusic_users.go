package database

import (
	"database/sql"
	"log"

	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	"golang.org/x/oauth2"
)

// SaveYTMusicToken saves or updates YouTube Music OAuth token and user info
func SaveYTMusicToken(db *sql.DB, token *oauth2.Token, user *model.GoogleUser) error {
	// トークンを暗号化
	encryptedAccessToken, err := utils.EncryptToken(token.AccessToken)
	if err != nil {
		log.Printf("Failed to encrypt access token: %v", err)
		return err
	}

	encryptedRefreshToken, err := utils.EncryptToken(token.RefreshToken)
	if err != nil {
		log.Printf("Failed to encrypt refresh token: %v", err)
		return err
	}

	_, err = db.Exec(`
        INSERT INTO ytmusic_users (id, email, access_token, refresh_token, token_expiration, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            email = EXCLUDED.email,
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            token_expiration = EXCLUDED.token_expiration,
            updated_at = NOW()`,
		user.ID, user.Email, encryptedAccessToken, encryptedRefreshToken, token.Expiry.Unix())
	if err != nil {
		return err
	}
	return nil
}

// UpdateYTMusicToken updates YouTube Music access token
func UpdateYTMusicToken(db *sql.DB, token *oauth2.Token, userID string) error {
	// アクセストークンを暗号化
	encryptedAccessToken, err := utils.EncryptToken(token.AccessToken)
	if err != nil {
		log.Printf("Failed to encrypt access token: %v", err)
		return err
	}

	_, err = db.Exec(`
        UPDATE ytmusic_users SET access_token = $1, token_expiration = $2, updated_at = NOW()
        WHERE id = $3`,
		encryptedAccessToken, token.Expiry.Unix(), userID)
	if err != nil {
		return err
	}
	return nil
}

// GetYTMusicUser retrieves YouTube Music user with decrypted tokens
func GetYTMusicUser(db *sql.DB, userID string) (*model.YTMusicUser, error) {
	var user model.YTMusicUser
	var encryptedAccessToken, encryptedRefreshToken string

	err := db.QueryRow(`
        SELECT id, email, access_token, refresh_token, token_expiration FROM ytmusic_users
        WHERE id = $1`, userID).Scan(&user.ID, &user.Email, &encryptedAccessToken, &encryptedRefreshToken, &user.TokenExpiration)
	if err != nil {
		return nil, err
	}

	// トークンを復号化
	user.AccessToken, err = utils.DecryptToken(encryptedAccessToken)
	if err != nil {
		log.Printf("Failed to decrypt access token: %v", err)
		return nil, err
	}

	user.RefreshToken, err = utils.DecryptToken(encryptedRefreshToken)
	if err != nil {
		log.Printf("Failed to decrypt refresh token: %v", err)
		return nil, err
	}

	return &user, nil
}
