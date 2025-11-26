package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetArtistAlbums(token *oauth2.Token, artistID string) ([]spotify.SimpleAlbum, error) {
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	var allAlbums []spotify.SimpleAlbum
	options := []spotify.RequestOption{spotify.Limit(50)}

	// 最初のページを取得
	albumsPage, err := client.GetArtistAlbums(ctx, spotify.ID(artistID), nil, options...)
	if err != nil {
		return nil, WrapSpotifyError(err)
	}

	// アルバムをallAlbumsに追加
	allAlbums = append(allAlbums, albumsPage.Albums...)

	// 次のページがある間はループして取得
	for albumsPage.Next != "" {
		// 次のページを取得
		err = client.NextPage(ctx, albumsPage)
		if err != nil {
			return nil, WrapSpotifyError(err)
		}
		allAlbums = append(allAlbums, albumsPage.Albums...)
	}

	return allAlbums, nil
}

func GetAlbumTracks(token *oauth2.Token, albumID string) ([]spotify.SimpleTrack, error) {
	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	var allTracks []spotify.SimpleTrack
	options := []spotify.RequestOption{spotify.Limit(50)}

	// 最初のページを取得
	tracksPage, err := client.GetAlbumTracks(ctx, spotify.ID(albumID), options...)
	if err != nil {
		return nil, WrapSpotifyError(err)
	}

	// トラックをallTracksに追加
	allTracks = append(allTracks, tracksPage.Tracks...)

	// 次のページがある間はループして取得
	for tracksPage.Next != "" {
		// `NextPage`メソッドで次のページを取得
		err = client.NextPage(ctx, tracksPage)
		if err != nil {
			return nil, WrapSpotifyError(err)
		}
		allTracks = append(allTracks, tracksPage.Tracks...)
	}

	return allTracks, nil
}
