package database

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pp-develop/music-timer-api/model"
	"golang.org/x/oauth2"
)

func SaveAccessToken(db *sql.DB, token *oauth2.Token, id string) error {
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

func UpdateAccessToken(db *sql.DB, token *oauth2.Token, id string) error {
	_, err := db.Exec(`
        UPDATE users SET access_token = $1, updated_at = NOW()
        WHERE id = $2`,
		token.AccessToken, id)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(db *sql.DB, id string) (model.User, error) {
	var user model.User

	err := db.QueryRow(`
        SELECT id, access_token, refresh_token, token_expiration, updated_at FROM users
        WHERE id = $1`, id).Scan(&user.Id, &user.AccessToken, &user.RefreshToken, &user.TokenExpiration, &user.UpdateAt)
	if err != nil {
		return user, err
	}

	return user, nil
}

func SaveFavoriteTrack(db *sql.DB, id string, track model.Track) error {
	// Track構造体をJSONにエンコード
	favoriteTrackJSON, err := json.Marshal(track)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO users (id, favorite_track, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (id) DO UPDATE SET
            favorite_track = EXCLUDED.favorite_track,
            updated_at = NOW()`,
		id, favoriteTrackJSON)
	if err != nil {
		return err
	}
	return nil
}

func GetFavoriteTracks(db *sql.DB, id string) ([]model.Track, error) {
	var tracksJSON string
	var tracks []model.Track

	// データベースからJSON文字列を取得
	err := db.QueryRow(
		"SELECT favorite_track FROM users WHERE id = $1", id).Scan(&tracksJSON)
	if err != nil {
		return nil, err
	}

	// JSON文字列を[]model.Trackにデコード
	err = json.Unmarshal([]byte(tracksJSON), &tracks)
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func AddFavoriteTrack(db *sql.DB, userId string, newTrack model.Track) error {
	var existingTracks []model.Track

	// 既存のJSONデータを取得
	var trackJSON *string
	err := db.QueryRow(`
        SELECT favorite_track FROM users
        WHERE id = $1`, userId).Scan(&trackJSON)
	if err != nil {
		return err
	}

	// JSONデータをデコード
	if trackJSON != nil && *trackJSON != "" { // trackJSON が NULL かどうかをチェック
		err = json.Unmarshal([]byte(*trackJSON), &existingTracks)
		if err != nil {
			return err
		}
	}

	// 新しいトラック情報を追加
	existingTracks = append(existingTracks, newTrack)

	// 更新されたトラックリストをJSONにエンコード
	updatedTrackJSON, err := json.Marshal(existingTracks)
	if err != nil {
		return err
	}

	// JSONカラムを更新
	_, err = db.Exec(`
        UPDATE users SET favorite_track = $1, updated_at = NOW()
        WHERE id = $2`, updatedTrackJSON, userId)
	if err != nil {
		return err
	}

	return nil
}

func UpdateUserUpdateAt(db *sql.DB, userId string, updatedAt time.Time) error {
	_, err := db.Exec(`
        UPDATE users SET updated_at = $1
        WHERE id = $2`,
		updatedAt.Format(time.RFC3339), userId)
	if err != nil {
		return err
	}
	return nil
}

func ClearFavoriteTracks(db *sql.DB, userId string) error {
	_, err := db.Exec(`
        UPDATE users SET favorite_track = '[]', updated_at = NOW()
        WHERE id = $1`, userId)
	if err != nil {
		return err
	}
	return nil
}
