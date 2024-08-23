package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

func GetArtistAlbums(artistID string) (*spotify.SimpleAlbumPage, error) {
	opts := []spotify.RequestOption{
		spotify.Market("JP"),
	}

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	albums, err := client.GetArtistAlbums(context.Background(), spotify.ID(artistID), nil, opts...)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func GetAlbumTracks(albumID string) (*spotify.SimpleTrackPage, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	tracks, err := client.GetAlbumTracks(context.Background(), spotify.ID(albumID))
	if err != nil {
		return nil, err
	}
	return tracks, nil
}
