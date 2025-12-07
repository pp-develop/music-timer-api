package auth

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

// AuthzNative generates SoundCloud authorization URL for native applications
func AuthzNative(c *gin.Context) (string, error) {
	client := soundcloud.NewClient()
	redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI_NATIVE")

	// For native apps, we don't need CSRF protection with state
	// as the response is returned directly via deep link
	url := client.GetAuthURL(redirectURI, "native")

	return url, nil
}

// CallbackNative handles SoundCloud OAuth callback for native applications
func CallbackNative(c *gin.Context) (*utils.TokenPair, error) {
	code := c.Query("code")

	db, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	// Exchange code for token
	client := soundcloud.NewClient()
	redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI_NATIVE")

	tokenResp, err := client.ExchangeCode(code, redirectURI)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := client.GetMe(tokenResp.AccessToken)
	if err != nil {
		return nil, err
	}

	// Calculate token expiration
	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		slog.Warn("token expiration not provided by API, using default 1 hour")
		expiresIn = 3600 // Default 1 hour
	}

	// Save user info to database
	user := &model.SoundCloudUser{
		Id:              strconv.Itoa(userInfo.ID),
		Username:        userInfo.Username,
		AccessToken:     tokenResp.AccessToken,
		RefreshToken:    tokenResp.RefreshToken,
		TokenExpiration: int(time.Now().Add(time.Duration(expiresIn) * time.Second).Unix()),
	}

	err = database.CreateOrUpdateSoundCloudUser(db, user)
	if err != nil {
		return nil, err
	}

	// Generate JWT token pair
	jti := uuid.New().String()
	tokenPair, err := utils.GenerateTokenPair(user.Id, "soundcloud", jti)
	if err != nil {
		return nil, err
	}

	// Save refresh token to database
	jwtExpiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	err = database.SaveRefreshToken(db, jti, user.Id, jwtExpiresAt)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func RefreshAccessToken(c *gin.Context) (*utils.TokenPair, error) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		return nil, model.ErrInvalidRequest
	}

	// Validate refresh token
	userID, service, jti, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Verify this is a SoundCloud token
	if service != "soundcloud" {
		return nil, model.ErrInvalidRefreshToken
	}

	db, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	// Check if refresh token is valid in database
	valid, err := database.IsRefreshTokenValid(db, jti)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, model.ErrInvalidRefreshToken
	}

	// Generate new token pair
	newJTI := uuid.New().String()
	tokenPair, err := utils.GenerateTokenPair(userID, service, newJTI)
	if err != nil {
		return nil, err
	}

	// Revoke old refresh token
	err = database.RevokeRefreshToken(db, jti)
	if err != nil {
		return nil, err
	}

	// Save new refresh token
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30 days
	err = database.SaveRefreshToken(db, newJTI, userID, expiresAt)
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}
