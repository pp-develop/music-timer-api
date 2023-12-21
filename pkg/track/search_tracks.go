package track

import (
	"net/url"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	spotifylibrary "github.com/zmb3/spotify/v2"
)

func SearchTracks() error {
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

func ValidateTime(msec int) bool {
	// 1minute = 60000ms
	oneminuteToMsec := 60000

	// 任意の分数 +- ALLOWANCE_MSEC を許容する
	// 下記の値だと+-5秒を許容するため、n分55秒~n分05秒の曲を許容する
	allowanceMsec := 5000

	for minute := 1; minute <= 8; minute++ {
		allowanceMsecMin := minute*oneminuteToMsec - allowanceMsec
		allowanceMsecMax := minute*oneminuteToMsec + allowanceMsec
		if msec >= allowanceMsecMin &&
			msec <= allowanceMsecMax {
			return true
		}
	}
	return false
}

func IsIsrcJp(isrc string) bool {
	return strings.Contains(isrc, "JP")
}
