package track

import (
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
	existNextUrl := true
	count := 0

	for existNextUrl {
		err := spotify.NextSearchTracks(items)
		if err != nil {
			return err
		}

		err = SaveTracks(items)
		if err != nil {
			return err
		}
		// 1000件trackを取得したら、items.Tracks.Nextが空になる想定だが、
		//　items.Tracks.Nextが空にならないので暫定対応
		if count == 1 {
			existNextUrl = false
		}
		if items.Tracks.Offset == 950 {
			count++
		}
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
