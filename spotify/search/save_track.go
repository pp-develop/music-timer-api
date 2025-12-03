package search

import (
	"database/sql"
	"log"
	"runtime"
	"strings"
	"time"

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
	start := time.Now()
	log.Printf("[SaveTracks] Start - market: %s", Market)

	c.BindJSON(&requestBody)
	if requestBody.Market != "" {
		Market = strings.ToUpper(requestBody.Market)
	}

	tracks, err := spotify.SearchTracks(Market)
	if err != nil {
		log.Printf("[SaveTracks] Error fetching from Spotify API: %v", err)
		return err
	}
	log.Printf("[SaveTracks] Fetched %d tracks from Spotify API", len(tracks))

	savedCount, err := saveTracks(db, tracks, true)
	if err != nil {
		log.Printf("[SaveTracks] Error saving to DB: %v", err)
		return err
	}

	log.Printf("[SaveTracks] Complete - fetched: %d, saved: %d, duration: %v", len(tracks), savedCount, time.Since(start))

	// 大量データ処理後にGCを実行してメモリを解放
	runtime.GC()

	return nil
}

func saveTracks(db *sql.DB, tracks []spotifylibrary.FullTrack, validate bool) (int, error) {
	// バリデーション済みトラックをフィルタリング
	validTracks := make([]spotifylibrary.FullTrack, 0, len(tracks))
	for _, item := range tracks {
		if validate && !validateTrack(item) {
			continue
		}
		validTracks = append(validTracks, item)
	}

	// バッチ保存（1回のDB呼び出しで全件保存）
	err := database.SaveTracksBatch(db, validTracks)
	return len(validTracks), err
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
