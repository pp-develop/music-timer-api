package auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

// CheckAuth checks authentication status without logging errors for unauthenticated users
// Use this for status check endpoints where unauthenticated is a valid state
func CheckAuth(c *gin.Context) (*model.SoundCloudUser, error) {
	db, ok := utils.GetDB(c)
	if !ok {
		slog.Error("failed to get DB instance")
		return nil, model.ErrFailedGetDB
	}

	userId, err := utils.GetUserID(c)
	if err != nil {
		// Not logged in - this is expected for status check, no error log
		return nil, model.ErrFailedGetSession
	}

	return getUserWithTokenRefresh(db, userId)
}

// GetAuth returns authenticated SoundCloud user
// Use this for protected endpoints where authentication is required
func GetAuth(c *gin.Context) (*model.SoundCloudUser, error) {
	db, ok := utils.GetDB(c)
	if !ok {
		slog.Error("failed to get DB instance")
		return nil, model.ErrFailedGetDB
	}

	userId, err := utils.GetUserID(c)
	if err != nil {
		slog.Error("failed to get user ID", slog.Any("error", err))
		return nil, model.ErrFailedGetSession
	}

	return getUserWithTokenRefresh(db, userId)
}

// getUserWithTokenRefresh retrieves user and refreshes SoundCloud token if expired
func getUserWithTokenRefresh(db *sql.DB, userId string) (*model.SoundCloudUser, error) {
	user, err := database.GetSoundCloudUser(db, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			// Session exists but user not in DB - treat as unauthenticated
			return nil, model.ErrFailedGetSession
		}
		slog.Error("failed to get user from database", slog.Any("error", err))
		return nil, err
	}

	// Check if token is expired
	if time.Now().Unix() > int64(user.TokenExpiration) {
		// Try to refresh token
		client := soundcloud.NewClient()
		tokenResp, err := client.RefreshToken(user.RefreshToken)
		if err != nil {
			slog.Error("failed to refresh token", slog.Any("error", err))
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Update tokens
		expiresIn := tokenResp.ExpiresIn
		if expiresIn == 0 {
			slog.Warn("token expiration not provided by API on refresh, using default 1 hour")
			expiresIn = 3600
		}

		err = database.UpdateSoundCloudUserTokens(
			db,
			userId,
			tokenResp.AccessToken,
			tokenResp.RefreshToken,
			int(time.Now().Add(time.Duration(expiresIn)*time.Second).Unix()),
		)
		if err != nil {
			return nil, err
		}

		user.AccessToken = tokenResp.AccessToken
		user.RefreshToken = tokenResp.RefreshToken
		user.TokenExpiration = int(time.Now().Add(time.Duration(expiresIn) * time.Second).Unix())
	}

	return user, nil
}
