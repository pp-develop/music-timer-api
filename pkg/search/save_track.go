package search

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

func SaveTracks() error {
	items, err := spotify.SearchTracks()
	if err != nil {
		return err
	}

	err = saveTracks(items, true)
	if err != nil {
		return err
	}

	err = nextSearchTracks(items)
	if err != nil {
		return err
	}
	return nil
}

func saveTracks(tracks *spotifylibrary.SearchResult, validate bool) error {
	for _, item := range tracks.Tracks.Tracks {
		if validate && !validateTrack(&item) {
			continue
		}
		err := database.SaveTrack(&item)
		if err != nil {
			return err
		}
	}
	return nil
}

func nextSearchTracks(items *spotifylibrary.SearchResult) error {
	var prevOffset string

	for {
		err := spotify.NextSearchTracks(items)
		if err != nil {
			return err
		}

		err = saveTracks(items, true)
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

func validateTrack(track *spotifylibrary.FullTrack) bool {
	return isIsrcJp(track.ExternalIDs["isrc"])
}

func isIsrcJp(isrc string) bool {
	return strings.Contains(isrc, "JP")
}
