package track

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	apiSpotify "github.com/pp-develop/music-timer-api/api/spotify" // Assuming this path based on the dummy file creation
)

// GetTracksByGenre handles requests for tracks by genre.
func GetTracksByGenre(c *gin.Context) {
	genreName := c.Param("genre_name")

	log.Printf("Received request for genre: %s", genreName)

	// Call the new GetRecommendationsByGenre function
	recommendations, err := apiSpotify.GetRecommendationsByGenre(genreName, "") // Passing empty market for now
	if err != nil {
		log.Printf("Error calling GetRecommendationsByGenre: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get recommendations",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully fetched recommendations",
		"genre":   genreName,
		"track_count": len(recommendations.Tracks),
		// Optionally, you could include more details from recommendations.Tracks if needed
		// For example: "tracks": recommendations.Tracks (but be mindful of data size)
	})
}
