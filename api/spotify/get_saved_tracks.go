package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// GetSavedTracks retrieves all saved tracks for the authenticated user
func GetSavedTracks(ctx context.Context, token *oauth2.Token) ([]spotify.SavedTrack, error) {
	var allTracks []spotify.SavedTrack

	client := NewClientWithToken(ctx, token)

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
