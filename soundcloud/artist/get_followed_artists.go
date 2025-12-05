package artist

import (
	"fmt"

	"github.com/gin-gonic/gin"
	soundcloud "github.com/pp-develop/music-timer-api/api/soundcloud"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/soundcloud/auth"
)

// GetFollowedArtists retrieves the artists (users) that the SoundCloud user is following
func GetFollowedArtists(c *gin.Context) ([]model.Artists, error) {
	// ユーザー情報を取得（SoundCloudトークンの期限切れ時は自動リフレッシュ）
	user, err := auth.GetAuth(c)
	if err != nil {
		return nil, err
	}

	client := soundcloud.NewClient()
	followings, err := client.GetFollowings(user.AccessToken)
	if err != nil {
		return nil, err
	}

	// SoundCloud UserをSpotifyと同じmodel.Artistsに変換
	allArtists := extractArtistInfo(followings)

	return allArtists, nil
}

// extractArtistInfo converts SoundCloud users to model.Artists
func extractArtistInfo(users []soundcloud.SCUser) []model.Artists {
	artistInfos := []model.Artists{}
	for _, user := range users {
		artistInfos = append(artistInfos, model.Artists{
			ImageUrl: user.AvatarURL,
			Id:       fmt.Sprintf("%d", user.ID),
			Name:     user.Username,
		})
	}
	return artistInfos
}
