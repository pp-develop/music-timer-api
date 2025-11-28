package database

import (
	"database/sql"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

func CreateOrUpdateSoundCloudUser(db *sql.DB, user *model.SoundCloudUser) error {
	_, err := db.Exec(`
        INSERT INTO soundcloud_users (id, username, access_token, refresh_token, token_expiration, session, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET
            username = EXCLUDED.username,
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            token_expiration = EXCLUDED.token_expiration,
            session = EXCLUDED.session,
            updated_at = NOW()`,
		user.Id, user.Username, user.AccessToken, user.RefreshToken, user.TokenExpiration, user.Session)
	return err
}

func GetSoundCloudUser(db *sql.DB, userId string) (*model.SoundCloudUser, error) {
	var user model.SoundCloudUser
	err := db.QueryRow(`
        SELECT id, username, access_token, refresh_token, token_expiration, session
        FROM soundcloud_users WHERE id = $1`, userId).Scan(
		&user.Id, &user.Username, &user.AccessToken, &user.RefreshToken, &user.TokenExpiration, &user.Session)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetSoundCloudUserBySession(db *sql.DB, session string) (*model.SoundCloudUser, error) {
	var user model.SoundCloudUser
	err := db.QueryRow(`
        SELECT id, username, access_token, refresh_token, token_expiration, session
        FROM soundcloud_users WHERE session = $1`, session).Scan(
		&user.Id, &user.Username, &user.AccessToken, &user.RefreshToken, &user.TokenExpiration, &user.Session)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateSoundCloudUserSession(db *sql.DB, userId, session string) error {
	_, err := db.Exec(`
        UPDATE soundcloud_users SET session = $1, updated_at = NOW()
        WHERE id = $2`, session, userId)
	return err
}

func DeleteSoundCloudUserSession(db *sql.DB, session string) error {
	_, err := db.Exec(`
        UPDATE soundcloud_users SET session = NULL, updated_at = NOW()
        WHERE session = $1`, session)
	return err
}

func UpdateSoundCloudUserTokens(db *sql.DB, userId, accessToken, refreshToken string, expiration int) error {
	_, err := db.Exec(`
        UPDATE soundcloud_users
        SET access_token = $1, refresh_token = $2, token_expiration = $3, updated_at = NOW()
        WHERE id = $4`,
		accessToken, refreshToken, expiration, userId)
	return err
}

func IncrementSoundCloudPlaylistCount(db *sql.DB, userId string) error {
	_, err := db.Exec(`
        UPDATE soundcloud_users
        SET playlist_count = playlist_count + 1, updated_at = NOW()
        WHERE id = $1`, userId)
	return err
}

func GetSoundCloudTracksUpdatedAt(db *sql.DB, userId string) (time.Time, error) {
	var updatedAt time.Time
	err := db.QueryRow(`
        SELECT updated_at FROM soundcloud_favorite_tracks WHERE user_id = $1`, userId).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	return updatedAt, nil
}
