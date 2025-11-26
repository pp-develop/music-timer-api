package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetSavedTracks(token *oauth2.Token) ([]spotify.SavedTrack, error) {
	var allTracks []spotify.SavedTrack

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient, spotify.WithRetry(true))

	tracksPage, err := client.CurrentUsersTracks(ctx, spotify.Limit(50))
	if err != nil {
		return nil, WrapSpotifyError(err)
	}
	allTracks = append(allTracks, tracksPage.Tracks...)

	// 次のページがある間はループして取得
	for tracksPage.Next != "" {
		// 次のページを取得
		err = client.NextPage(ctx, tracksPage)
		if err != nil {
			return nil, WrapSpotifyError(err)
		}
		allTracks = append(allTracks, tracksPage.Tracks...)
	}

	return allTracks, nil
}
