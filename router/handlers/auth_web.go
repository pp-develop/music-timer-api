package handlers

import (
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/auth"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

// GetAuthzURLWeb returns the Spotify authorization URL for web authentication
func GetAuthzURLWeb(c *gin.Context) {
	url, err := auth.SpotifyAuthzWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}

// CallbackWeb handles the Spotify OAuth callback for web authentication
func CallbackWeb(c *gin.Context) {
	err := godotenv.Load()
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	err = auth.SpotifyCallbackWeb(c)
	if err != nil {
		logger.LogError(err)
		c.Redirect(http.StatusSeeOther, os.Getenv("AUTHZ_ERROR_URL"))
	} else {
		c.Redirect(http.StatusMovedPermanently, os.Getenv("AUTHZ_SUCCESS_URL"))
	}
}

// GetAuthStatusWeb returns the authentication status for web users
func GetAuthStatusWeb(c *gin.Context) {
	user, err := auth.Auth(c)

	if err == model.ErrFailedGetSession {
		logger.LogError(err)
		c.Status(http.StatusSeeOther)
	} else if err != nil {
		logger.LogError(err)
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, gin.H{"country": user.Country})
	}
}

// DeleteSession deletes the user's session
func DeleteSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userId")
	session.Save()
	c.Status(http.StatusOK)
}
