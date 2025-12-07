package database

import (
	"database/sql"
	"time"
)

// SaveSoundCloudRefreshToken saves a new refresh token to the database
func SaveSoundCloudRefreshToken(db *sql.DB, jti, userID string, expiresAt time.Time) error {
	query := `
		INSERT INTO soundcloud_jwt_refresh_tokens (jti, user_id, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, false)
	`

	_, err := db.Exec(query, jti, userID, expiresAt)
	return err
}

// IsSoundCloudRefreshTokenValid checks if a refresh token exists and is valid
func IsSoundCloudRefreshTokenValid(db *sql.DB, jti string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM soundcloud_jwt_refresh_tokens
		WHERE jti = $1
		AND revoked = false
		AND expires_at > CURRENT_TIMESTAMP
	`

	err := db.QueryRow(query, jti).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// RevokeSoundCloudRefreshToken marks a refresh token as revoked
func RevokeSoundCloudRefreshToken(db *sql.DB, jti string) error {
	query := `
		UPDATE soundcloud_jwt_refresh_tokens
		SET revoked = true
		WHERE jti = $1
	`

	_, err := db.Exec(query, jti)
	return err
}

// RevokeAllSoundCloudUserRefreshTokens revokes all refresh tokens for a user
func RevokeAllSoundCloudUserRefreshTokens(db *sql.DB, userID string) error {
	query := `
		UPDATE soundcloud_jwt_refresh_tokens
		SET revoked = true
		WHERE user_id = $1 AND revoked = false
	`

	_, err := db.Exec(query, userID)
	return err
}

// CleanupExpiredSoundCloudTokens removes expired refresh tokens from the database
func CleanupExpiredSoundCloudTokens(db *sql.DB) error {
	query := `
		DELETE FROM soundcloud_jwt_refresh_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query)
	return err
}
