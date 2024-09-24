package spotify

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func GetArtistAlbums(artistID string) ([]spotify.SimpleAlbum, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	var allAlbums []spotify.SimpleAlbum

	// 最初のページを取得
	albumsPage, err := client.GetArtistAlbums(ctx, spotify.ID(artistID), nil)
	if err != nil {
		return nil, err
	}

	// アルバムをallAlbumsに追加
	allAlbums = append(allAlbums, albumsPage.Albums...)

	// 次のページがある間はループして取得
	for albumsPage.Next != "" {
		// 次のページを取得
		err = client.NextPage(ctx, albumsPage)
		if err != nil {
			return nil, err
		}
		allAlbums = append(allAlbums, albumsPage.Albums...)
	}

	return allAlbums, nil
}

func GetAlbumTracks(albumID string) ([]spotify.SimpleTrack, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	var allTracks []spotify.SimpleTrack
	options := []spotify.RequestOption{spotify.Limit(50)}

	// 最初のページを取得
	tracksPage, err := client.GetAlbumTracks(ctx, spotify.ID(albumID), options...)
	if err != nil {
		return nil, err
	}

	// トラックをallTracksに追加
	allTracks = append(allTracks, tracksPage.Tracks...)

	// 次のページがある間はループして取得
	for tracksPage.Next != "" {
		// `NextPage`メソッドで次のページを取得
		err = client.NextPage(ctx, tracksPage)
		if err != nil {
			return nil, err
		}
		allTracks = append(allTracks, tracksPage.Tracks...)
	}

	return allTracks, nil
}
