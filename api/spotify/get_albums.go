package spotify

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func GetArtistAlbums(artistID string) (*spotify.SimpleAlbumPage, error) {
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
	albums, err := client.GetArtistAlbums(context.Background(), spotify.ID(artistID), nil)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func GetAlbumTracks(albumID string) (*spotify.SimpleTrackPage, error) {
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
	tracks, err := client.GetAlbumTracks(context.Background(), spotify.ID(albumID))
	if err != nil {
		return nil, err
	}
	return tracks, nil
}
