package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/urfave/cli/v2"
	spotifylibrary "github.com/zmb3/spotify/v2"
)

// 1 minute ＝ 60000 msecond
const ONEMINUTE_TO_MSEC = 60000

// 任意の分数 +- ALLOWANCE_MSEC を許容する
// 下記の値だと+-5秒を許容するため、n分55秒~n分05秒の曲を許容する
const ALLOWANCE_MSEC = 5000

func main() {
	app := &cli.App{
		Name:  "save_track",
		Usage: "search spotify tracks to DB save",
		Action: func(*cli.Context) error {
			fmt.Println("start save track")
			invoke()
			fmt.Println("end save track")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func invoke() {
	searchResult := spotify.SearchTracks()
	saveTracks(searchResult)
	NextSearchTracks(searchResult)
}

func saveTracks(tracks *spotifylibrary.SearchResult) {
	for _, item := range tracks.Tracks.Tracks {
		if validateTrack(&item) {
			database.SaveTrack(&item)
		}
	}
}

func NextSearchTracks(tracks *spotifylibrary.SearchResult) {
	searchSuccess := true
	var items *spotifylibrary.SearchResult

	for searchSuccess {
		searchSuccess, items = spotify.NextSearchTracks(tracks)
		if searchSuccess {
			saveTracks(items)
		}
	}
}

func validateTrack(track *spotifylibrary.FullTrack) bool {
	if isIsrcJp(track.ExternalIDs["isrc"]) && validateTime(track.Duration) {
		return true
	}
	return false
}

func validateTime(msec int) bool {
	for minute := 1; minute <= 8; minute++ {
		if msec >= minute*ONEMINUTE_TO_MSEC-ALLOWANCE_MSEC &&
			msec <= minute*ONEMINUTE_TO_MSEC+ALLOWANCE_MSEC {
			return true
		}
	}
	return false
}

func isIsrcJp(isrc string) bool {
	return isrc[0:2] == "JP"
}
