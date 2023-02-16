package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetSavedTracks(token *oauth2.Token) (*spotify.SavedTrackPage, error) {
	var tracks *spotify.SavedTrackPage

	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	tracks, err := client.CurrentUsersTracks(ctx)
	if err != nil {
		return tracks, err
	}

	return tracks, nil
}

func GetNextSavedTrakcs(token *oauth2.Token, track *spotify.SavedTrackPage) error {

	ctx := context.Background()
	httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient)

	err := client.NextPage(ctx, track)
	if err != nil {
		return err
	}

	return nil
}
