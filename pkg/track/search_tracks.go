package track

import (
	"net/url"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	spotifylibrary "github.com/zmb3/spotify/v2"
)

type RequestJson struct {
	IncludeFavoriteArtists bool `json:"includeFavoriteArtists"`
}

func SearchAndSaveTracks() error {
	items, err := spotify.SearchTracks()
	if err != nil {
		return err
	}

	err = SaveTracks(items)
	if err != nil {
		return err
	}

	err = NextSearchTracks(items)
	if err != nil {
		return err
	}
	return nil
}

func SaveTracks(tracks *spotifylibrary.SearchResult) error {
	for _, item := range tracks.Tracks.Tracks {
		if !ValidateTrack(&item) {
			continue
		}
		err := database.SaveTrack(&item)
		if err != nil {
			return err
		}
	}
	return nil
}

func NextSearchTracks(items *spotifylibrary.SearchResult) error {
	var prevOffset string

	for {
		err := spotify.NextSearchTracks(items)
		if err != nil {
			return err
		}

		err = SaveTracks(items)
		if err != nil {
			return err
		}

		if items.Tracks.Next == "" {
			break
		}

		parsedURL, err := url.Parse(items.Tracks.Next)
		if err != nil {
			return err
		}
		queryParams := parsedURL.Query()
		currentOffset := queryParams.Get("offset")

		// 同じoffsetが2回続いたらループを終了
		if currentOffset == prevOffset {
			break
		}

		prevOffset = currentOffset // 現在のoffsetを保存
	}

	return nil
}

func ValidateTrack(track *spotifylibrary.FullTrack) bool {
	return IsIsrcJp(track.ExternalIDs["isrc"])
}

func IsIsrcJp(isrc string) bool {
	return strings.Contains(isrc, "JP")
}
