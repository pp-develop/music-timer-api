package apiSpotify

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth" // Alias for spotify/v2/auth
	"golang.org/x/oauth2/clientcredentials"
)

// GetRecommendationsByGenreFunc is a function variable that can be swapped out for testing.
// It defaults to the actual implementation.
var GetRecommendationsByGenreFunc = getRecommendationsByGenreInternal

// GetRecommendationsByGenre is the public function that clients will call.
// It now calls the function variable, allowing it to be mocked.
func GetRecommendationsByGenre(genreName string, market string) (*spotify.Recommendations, error) {
	return GetRecommendationsByGenreFunc(genreName, market)
}

// getRecommendationsByGenreInternal is the original implementation.
func getRecommendationsByGenreInternal(genreName string, market string) (*spotify.Recommendations, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		// It's often better to return the error rather than fatally exiting in a library function
		// However, if SPOTIFY_ID and SPOTIFY_SECRET are critical and expected to be in .env,
		// this behavior might be acceptable, or a more specific error could be returned.
		// For now, we'll proceed, assuming they might be set in the environment directly.
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		log.Printf("Error getting token: %v", err)
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	seeds := spotify.Seeds{
		Genres: []string{genreName},
	}

	// Setup recommendation options
	options := []spotify.RequestOption{
		spotify.Limit(50), // Default limit
	}

	if market != "" {
		options = append(options, spotify.Market(market))
	}

	recommendations, err := client.GetRecommendations(ctx, seeds, nil, options...)
	if err != nil {
		log.Printf("Error getting recommendations: %v", err)
		return nil, err
	}

	return recommendations, nil
}
