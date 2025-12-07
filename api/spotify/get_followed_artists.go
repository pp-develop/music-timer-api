package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// GetFollowedArtists retrieves all artists followed by the current user
func GetFollowedArtists(ctx context.Context, token *oauth2.Token) ([]spotify.FullArtist, error) {
	var allArtists []spotify.FullArtist

	client := NewClientWithToken(ctx, token)

	artistPage, err := client.CurrentUsersFollowedArtists(ctx, spotify.Limit(50))
	if err != nil {
		return nil, WrapSpotifyError(err)
	}
	allArtists = append(allArtists, artistPage.Artists...)

	// 次のページがある間はループして取得
	for {
		if artistPage.Cursor.After == "" {
			break
		}

		// 次のページを取得
		artistPage, err = client.CurrentUsersFollowedArtists(ctx, spotify.Limit(50), spotify.After(artistPage.Cursor.After))
		if err != nil {
			return nil, WrapSpotifyError(err)
		}
		allArtists = append(allArtists, artistPage.Artists...)
	}

	return allArtists, nil
}
