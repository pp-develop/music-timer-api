package artist

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// GetFollowedArtists は、Spotifyユーザーがフォローしたアーティストを取得します。
func GetFollowedArtists(c *gin.Context) ([]model.Artists, error) {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return nil, model.ErrFailedGetSession
	}
	userId := v.(string)

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	user, err := database.GetUser(dbInstance, userId)
	if err != nil {
		return nil, err
	}

	artistsPage, err := spotifyApi.GetFollowedArtists(&oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
	})
	if err != nil {
		return nil, err
	}

	// 最初のページのアーティストを取得し、IDと名前を抽出
	allArtists := extractArtistInfo(artistsPage.Artists)

	// 次のページが存在する場合、追加のアーティストを取得
	allArtists, err = fetchNextArtists(
		&oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		}, artistsPage, allArtists)
	if err != nil {
		return nil, err
	}

	return allArtists, nil
}

func fetchNextArtists(token *oauth2.Token, currentPage *spotify.FullArtistCursorPage, allArtists []model.Artists) ([]model.Artists, error) {
	for currentPage.Cursor.After != "" {
		var err error
		currentPage, err = spotifyApi.GetAfterFollowedArtists(token, currentPage.Cursor.After)
		if err != nil {
			return nil, err
		}

		// 新しいページのアーティストのIDと名前を追加
		allArtists = append(allArtists, extractArtistInfo(currentPage.Artists)...)
	}

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
