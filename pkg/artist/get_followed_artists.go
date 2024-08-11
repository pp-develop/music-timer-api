package artist

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	spotifyApi "github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type ArtistInfo struct {
	ID       string
	Name     string
	ImageUrl string
}

// GetFollowedArtists は、Spotifyユーザーがフォローしたアーティストを取得します。
func GetFollowedArtists(c *gin.Context) ([]ArtistInfo, error) {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return nil, model.ErrFailedGetSession
	}
	userId := v.(string)
	user, err := database.GetUser(userId)
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

func fetchNextArtists(token *oauth2.Token, currentPage *spotify.FullArtistCursorPage, allArtists []ArtistInfo) ([]ArtistInfo, error) {
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
func extractArtistInfo(artists []spotify.FullArtist) []ArtistInfo {
	var artistInfos []ArtistInfo
	for _, artist := range artists {
		if len(artist.Images) > 0 {
			artistInfos = append(artistInfos, ArtistInfo{
				ImageUrl: artist.Images[0].URL,
				ID:       string(artist.ID),
				Name:     artist.Name})
		} else {
			artistInfos = append(artistInfos, ArtistInfo{
				ImageUrl: "", // 画像がない場合のデフォルト値
				ID:       string(artist.ID),
				Name:     artist.Name})
		}
	}
	return artistInfos
}
