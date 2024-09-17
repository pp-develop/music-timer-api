package search

import (
	"database/sql"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	spotifylibrary "github.com/zmb3/spotify/v2"
)

var (
	Market string
)

var requestBody struct {
	Market string `json:"market"`
}

func SaveTracks(c *gin.Context, db *sql.DB) error {
	c.BindJSON(&requestBody)
	if requestBody.Market != "" {
		Market = strings.ToUpper(requestBody.Market)
	}

	items, err := spotify.SearchTracks(Market)
	if err != nil {
		return err
	}

	err = saveTracks(db, items, true)
	if err != nil {
		return err
	}

	err = nextSearchTracks(db, items)
	if err != nil {
		return err
	}
	return nil
}

func saveTracks(db *sql.DB, tracks *spotifylibrary.SearchResult, validate bool) error {
	for _, item := range tracks.Tracks.Tracks {
		if validate && !validateTrack(&item) {
			continue
		}
		err := database.SaveTrack(db, &item)
		if err != nil {
			return err
		}
	}
	return nil
}

func nextSearchTracks(db *sql.DB, items *spotifylibrary.SearchResult) error {
	var prevOffset string

	for {
		err := spotify.NextSearchTracks(items)
		if err != nil {
			return err
		}

		err = saveTracks(db, items, true)
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
	return isIsrcForMarket(track.ExternalIDs["isrc"])
}

func isIsrcForMarket(isrc string) bool {
	if Market == "" {
		return true
	}
	return strings.Contains(isrc, Market)
}
