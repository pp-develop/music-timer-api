package handlers

import (
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

// GetAuthzURLSoundCloud returns the SoundCloud authorization URL
func GetAuthzURLSoundCloud(c *gin.Context) {
	url, err := auth.SoundCloudAuthzWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackSoundCloud handles the SoundCloud OAuth callback
func CallbackSoundCloud(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	err = auth.SoundCloudCallbackWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

// GetAuthStatusSoundCloud returns the authentication status for SoundCloud users
func GetAuthStatusSoundCloud(c *gin.Context) {
	user, err := auth.GetSoundCloudAuthStatus(c)

	if err == model.ErrFailedGetSession {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"reason":        "session_expired",
		})
	} else if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": true,
			"username":      user.Username,
		})
	}
}

// DeleteSessionSoundCloud deletes the SoundCloud user's session
func DeleteSessionSoundCloud(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("sc_userId")
	session.Save()
	c.Status(http.StatusOK)
}
