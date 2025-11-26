package handler

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/auth"
	"github.com/pp-develop/music-timer-api/pkg/track"
	"github.com/pp-develop/music-timer-api/pkg/ytmusic"
	"github.com/pp-develop/music-timer-api/utils"
)

// YTMusicAuthzUrlWeb generates YouTube Music authorization URL for web
func YTMusicAuthzUrlWeb(c *gin.Context) {
	url, err := auth.YTMusicAuthzWeb(c)
	if err != nil {
		log.Printf("[ERROR] Failed to generate YTMusic auth URL: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

// YTMusicCallbackWeb handles YouTube Music OAuth callback
func YTMusicCallbackWeb(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	err := auth.YTMusicCallbackWeb(c, db)
	if err != nil {
		log.Printf("[ERROR] YTMusic callback failed: %v", err)
		c.Redirect(http.StatusSeeOther, "/ytmusic/auth/error")
		return
	}

	c.Redirect(http.StatusSeeOther, "/ytmusic/auth/success")
}

// YTMusicAuthStatus checks YouTube Music authentication status
func YTMusicAuthStatus(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("ytmusic_user_id")

	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	user, err := database.GetYTMusicUser(db, userID.(string))
	if err != nil {
		log.Printf("[ERROR] Failed to get YTMusic user: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"email":         user.Email,
	})
}

// SaveYTMusicFavoriteTracks saves user's YouTube Music liked songs
func SaveYTMusicFavoriteTracks(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	session := sessions.Default(c)
	userID := session.Get("ytmusic_user_id")
	if userID == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Create YTMusic client
	client := ytmusic.NewClient(db, userID.(string))

	// Get liked songs
	tracks, err := client.GetLikedSongs()
	if err != nil {
		log.Printf("[ERROR] Failed to get liked songs: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Save to database
	err = database.SaveYTMusicFavoriteTracks(db, userID.(string), tracks)
	if err != nil {
		log.Printf("[ERROR] Failed to save favorite tracks: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] Saved %d favorite tracks for user %s", len(tracks), userID)
	c.JSON(http.StatusOK, gin.H{"count": len(tracks)})
}

// GetYTMusicFavoriteTracks retrieves user's YouTube Music liked songs from DB
func GetYTMusicFavoriteTracks(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	session := sessions.Default(c)
	userID := session.Get("ytmusic_user_id")
	if userID == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	tracks, err := database.GetYTMusicFavoriteTracks(db, userID.(string))
	if err != nil {
		log.Printf("[ERROR] Failed to get favorite tracks: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tracks": tracks, "count": len(tracks)})
}

// CreateYTMusicPlaylist creates a YouTube Music playlist from favorites
func CreateYTMusicPlaylist(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	session := sessions.Default(c)
	userID := session.Get("ytmusic_user_id")
	if userID == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		DurationMinutes int      `json:"duration_minutes" binding:"required"`
		Title           string   `json:"title"`
		Description     string   `json:"description"`
		ArtistIDs       []string `json:"artist_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get favorite tracks from database
	ytTracks, err := database.GetYTMusicFavoriteTracks(db, userID.(string))
	if err != nil {
		log.Printf("[ERROR] Failed to get favorite tracks: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	if len(ytTracks) == 0 {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeNoFavoriteTracks,
		})
		return
	}

	// Convert YouTube tracks to generic Track format for algorithm
	tracks := make([]model.Track, len(ytTracks))
	for i, yt := range ytTracks {
		tracks[i] = model.Track{
			Uri:        yt.VideoID, // Use video_id as URI
			DurationMs: yt.DurationMs,
		}
	}

	// Use existing MakeTracks algorithm to find optimal combination
	specifyMs := req.DurationMinutes * 60 * 1000
	success, selectedTracks := track.MakeTracks(tracks, specifyMs)

	if !success {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeNotEnoughTracks,
		})
		return
	}

	// Extract video IDs from selected tracks
	videoIDs := make([]string, len(selectedTracks))
	for i, t := range selectedTracks {
		videoIDs[i] = t.Uri
	}

	// Create playlist title if not provided
	title := req.Title
	if title == "" {
		title = "Music Timer Playlist"
	}

	// Create YTMusic client and create playlist
	client := ytmusic.NewClient(db, userID.(string))
	playlistID, err := client.CreatePlaylist(title, req.Description, videoIDs)
	if err != nil {
		log.Printf("[ERROR] Failed to create playlist: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Save playlist to database
	err = database.SaveYTMusicPlaylist(db, playlistID, userID.(string))
	if err != nil {
		log.Printf("[WARN] Failed to save playlist to DB: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"playlist_id": playlistID,
		"track_count": len(videoIDs),
	})
}

// GetYTMusicPlaylists retrieves user's created YouTube Music playlists
func GetYTMusicPlaylists(c *gin.Context) {
	db, ok := utils.GetDB(c)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	session := sessions.Default(c)
	userID := session.Get("ytmusic_user_id")
	if userID == nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	playlists, err := database.GetAllYTMusicPlaylists(db, userID.(string))
	if err != nil {
		log.Printf("[ERROR] Failed to get playlists: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"playlists": playlists})
}

// DeleteYTMusicSession logs out the user from YouTube Music
func DeleteYTMusicSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("ytmusic_user_id")
	session.Delete("ytmusic_state")
	session.Save()
	c.Status(http.StatusOK)
}
