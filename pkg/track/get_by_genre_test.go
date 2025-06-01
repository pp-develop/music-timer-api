package track

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apiSpotify "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"
)

func TestGetTracksByGenre(t *testing.T) {
	// Store original GetRecommendationsByGenreFunc and restore it after the test
	originalGetRecommendationsByGenreFunc := apiSpotify.GetRecommendationsByGenreFunc
	defer func() {
		apiSpotify.GetRecommendationsByGenreFunc = originalGetRecommendationsByGenreFunc
	}()

	gin.SetMode(gin.TestMode)

	t.Run("Success Case", func(t *testing.T) {
		// Mock apiSpotify.GetRecommendationsByGenre
		apiSpotify.GetRecommendationsByGenreFunc = func(genreName string, market string) (*spotify.Recommendations, error) {
			return &spotify.Recommendations{
				Tracks: []spotify.FullTrack{
					{SimpleTrack: spotify.SimpleTrack{ID: "track1", Name: "Track 1"}},
					{SimpleTrack: spotify.SimpleTrack{ID: "track2", Name: "Track 2"}},
				},
			}, nil
		}

		router := gin.New()
		router.GET("/tracks/genre/:genre_name", GetTracksByGenre)

		req, _ := http.NewRequest(http.MethodGet, "/tracks/genre/pop", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Successfully fetched recommendations", response["message"])
		assert.Equal(t, "pop", response["genre"])
		// Note: The track_count will be float64 due to JSON unmarshalling into interface{}
		assert.Equal(t, 2.0, response["track_count"])
	})

	t.Run("Spotify API Error Case", func(t *testing.T) {
		// Mock apiSpotify.GetRecommendationsByGenre to return an error
		apiSpotify.GetRecommendationsByGenreFunc = func(genreName string, market string) (*spotify.Recommendations, error) {
			return nil, errors.New("Spotify API error")
		}

		router := gin.New()
		router.GET("/tracks/genre/:genre_name", GetTracksByGenre)

		req, _ := http.NewRequest(http.MethodGet, "/tracks/genre/rock", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "Failed to get recommendations", response["message"])
		assert.Equal(t, "Spotify API error", response["error"])
	})

	// Optional: Test for specific behavior with empty genre if desired,
	// though Spotify might handle it or return an error.
	// For now, the above cases cover success and direct API error.
	t.Run("Empty Genre Name Case (Still Hits Mock)", func(t *testing.T) {
		// This test primarily ensures the handler passes the genre name correctly,
		// and the mock intercepts it. The mock will return success here.
		apiSpotify.GetRecommendationsByGenreFunc = func(genreName string, market string) (*spotify.Recommendations, error) {
			// Assert that the genreName received by the mock is indeed empty if that's what we're testing
			assert.Equal(t, "", genreName, "Expected empty genre name to be passed to mock")
			return &spotify.Recommendations{
				Tracks: []spotify.FullTrack{}, // No tracks for empty genre
			}, nil
		}

		router := gin.New()
		router.GET("/tracks/genre/:genre_name", GetTracksByGenre)

		// Test with an effectively empty genre name in the path
		req, _ := http.NewRequest(http.MethodGet, "/tracks/genre/", nil)
		rr := httptest.NewRecorder()

		// Note: Gin might not route this as expected if :genre_name is required to be non-empty.
		// Let's assume for this test that an "empty" genre parameter might be passed if the route was /tracks/genre/:genre_name/*param
		// or if the parameter is optional. Given "/tracks/genre/:genre_name", an empty string here would mean the path is "/tracks/genre/"
		// which might not match. Let's use a placeholder like "unknown" and verify it's passed.
		// If we want to test a truly empty path segment, the routing itself would be the first point of failure.
		// For this test, let's assume "unknown" is a genre that results in 0 tracks.

		req_unknown, _ := http.NewRequest(http.MethodGet, "/tracks/genre/unknown", nil)
		rr_unknown := httptest.NewRecorder()
		router.ServeHTTP(rr_unknown, req_unknown)

		assert.Equal(t, http.StatusOK, rr_unknown.Code)
		var response_unknown map[string]interface{}
		err_unknown := json.Unmarshal(rr_unknown.Body.Bytes(), &response_unknown)
		assert.NoError(t, err_unknown)
		assert.Equal(t, "Successfully fetched recommendations", response_unknown["message"])
		assert.Equal(t, "unknown", response_unknown["genre"])
		assert.Equal(t, 0.0, response_unknown["track_count"])
	})
}
