package main

import (
	"log"
	"os"

	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/urfave/cli/v2"
	spotifylibrary "github.com/zmb3/spotify/v2"
)

func main() {
	app := &cli.App{
		Name:  "save_track",
		Usage: "search spotify tracks to DB save",
		Action: func(*cli.Context) error {
			log.Println("start save track")
			Invoke()
			log.Println("end save track")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Invoke() {
	searchResult, err := spotify.SearchTracks()
	if err != nil {
		log.Fatal(err)
	}

	err = SaveTracks(searchResult)
	if err != nil {
		log.Fatal(err)
	}

	err = NextSearchTracks(searchResult)
	if err != nil {
		log.Fatal(err)
	}
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

func NextSearchTracks(tracks *spotifylibrary.SearchResult) error {
	searchSuccess := true

	for searchSuccess {
		items, err := spotify.NextSearchTracks(tracks)
		if err != nil {
			return err
		}

		err = SaveTracks(items)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateTrack(track *spotifylibrary.FullTrack) bool {
	if IsIsrcJp(track.ExternalIDs["isrc"]) && ValidateTime(track.Duration) {
		return true
	}
	return false
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
	return isrc[0:2] == "JP"
}
