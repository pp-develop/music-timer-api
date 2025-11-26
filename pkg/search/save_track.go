package search

import (
	"database/sql"
	"log"
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

	tracks, err := spotify.SearchTracks(Market)
	if err != nil {
		return err
	}

	err = saveTracks(db, tracks, true)
	if err != nil {
		return err
	}

	log.Printf("Total number of tracks: %d\n", len(tracks))
	return nil
}

func saveTracks(db *sql.DB, tracks []spotifylibrary.FullTrack, validate bool) error {
	for _, item := range tracks {
		if validate && !validateTrack(item) {
			continue
		}
		err := database.SaveTrack(db, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveTracksForCLI(db *sql.DB, market string) error {
	if market != "" {
		Market = strings.ToUpper(market)
	}

	tracks, err := spotify.SearchTracks(Market)
	if err != nil {
		return err
	}

	err = saveTracks(db, tracks, true)
	if err != nil {
		return err
	}

	log.Printf("Total number of tracks: %d\n", len(tracks))
	return nil
}

func validateTrack(track spotifylibrary.FullTrack) bool {
	return isIsrcForMarket(track.ExternalIDs["isrc"])
}

func isIsrcForMarket(isrc string) bool {
	if Market == "" {
		return true
	}
	return strings.Contains(isrc, Market)
}
