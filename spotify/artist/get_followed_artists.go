package artist

import (
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/auth"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// GetFollowedArtists は、Spotifyユーザーがフォローしたアーティストを取得します。
func GetFollowedArtists(c *gin.Context) ([]model.Artists, error) {
	// ユーザー情報を取得（Spotifyトークンの期限切れ時は自動リフレッシュ）
	user, err := auth.GetUserWithValidToken(c)
	if err != nil {
		return nil, err
	}

	followedArtists, err := spotifyApi.GetFollowedArtists(&oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	// 最初のページのアーティストを取得し、IDと名前を抽出
	allArtists := extractArtistInfo(followedArtists)

	return allArtists, nil
}

// extractArtistInfo は、spotify.FullArtist のスライスから ArtistInfo のスライスを生成します。
func extractArtistInfo(artists []spotify.FullArtist) []model.Artists {
	artistInfos := []model.Artists{}
	for _, artist := range artists {
		if len(artist.Images) > 0 {
			artistInfos = append(artistInfos, model.Artists{
				ImageUrl: artist.Images[0].URL,
				Id:       string(artist.ID),
				Name:     artist.Name})
		} else {
			artistInfos = append(artistInfos, model.Artists{
				ImageUrl: "", // 画像がない場合のデフォルト値
				Id:       string(artist.ID),
				Name:     artist.Name})
		}
	}
	return artistInfos
}
