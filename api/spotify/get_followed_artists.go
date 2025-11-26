package spotify

import (
	"context"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetFollowedArtists(token *oauth2.Token) ([]spotify.FullArtist, error) {
	var allArtists []spotify.FullArtist

	ctx := context.Background()
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client := spotify.New(httpClient, spotify.WithRetry(true))

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
