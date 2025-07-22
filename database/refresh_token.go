package database

import (
	"database/sql"
	"time"
)

// RefreshToken represents a refresh token record in the database
type RefreshToken struct {
	JTI       string    `db:"jti"`
	UserID    string    `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
	Revoked   bool      `db:"revoked"`
}

// SaveRefreshToken saves a new refresh token to the database
func SaveRefreshToken(db *sql.DB, jti, userID string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (jti, user_id, expires_at, created_at, revoked)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, false)
	`
	
	_, err := db.Exec(query, jti, userID, expiresAt)
	return err
}

// IsRefreshTokenValid checks if a refresh token exists and is valid
func IsRefreshTokenValid(db *sql.DB, jti string) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM refresh_tokens 
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

// RevokeRefreshToken marks a refresh token as revoked
func RevokeRefreshToken(db *sql.DB, jti string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked = true 
		WHERE jti = $1
	`
	
	_, err := db.Exec(query, jti)
	return err
}

// RevokeAllUserRefreshTokens revokes all refresh tokens for a user
func RevokeAllUserRefreshTokens(db *sql.DB, userID string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked = true 
		WHERE user_id = $1 AND revoked = false
	`
	
	_, err := db.Exec(query, userID)
	return err
}

// CleanupExpiredTokens removes expired refresh tokens from the database
func CleanupExpiredTokens(db *sql.DB) error {
	query := `
		DELETE FROM refresh_tokens 
		WHERE expires_at < CURRENT_TIMESTAMP
	`
	
	_, err := db.Exec(query)
	return err
}