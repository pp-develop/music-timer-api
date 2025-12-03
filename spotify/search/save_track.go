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

func SaveTracks(c *gin.Context, db *sql.DB) error {
	start := time.Now()

	var requestBody struct {
		Market string `json:"market"`
	}
	c.BindJSON(&requestBody)

	market := strings.ToUpper(requestBody.Market)
	log.Printf("[SaveTracks] Start - market: %s", market)

	tracks, err := spotify.SearchTracks(market)
	if err != nil {
		log.Printf("[SaveTracks] Error fetching from Spotify API: %v", err)
		return err
	}
	log.Printf("[SaveTracks] Fetched %d tracks from Spotify API", len(tracks))

	savedCount, err := saveTracks(db, tracks, market, true)
	if err != nil {
		log.Printf("[SaveTracks] Error saving to DB: %v", err)
		return err
	}

	log.Printf("[SaveTracks] Complete - fetched: %d, saved: %d, duration: %v", len(tracks), savedCount, time.Since(start))

	// 大量データ処理後にGCを実行してメモリを解放
	runtime.GC()

	return nil
}

func saveTracks(db *sql.DB, tracks []spotifylibrary.FullTrack, market string, validate bool) (int, error) {
	// バリデーション済みトラックをフィルタリング
	validTracks := make([]spotifylibrary.FullTrack, 0, len(tracks))
	for _, item := range tracks {
		if validate && !validateTrack(item, market) {
			continue
		}
		validTracks = append(validTracks, item)
	}

	// バッチ保存（1回のDB呼び出しで全件保存）
	err := database.SaveTracksBatch(db, validTracks)
	return len(validTracks), err
}

func validateTrack(track spotifylibrary.FullTrack, market string) bool {
	return isIsrcForMarket(track.ExternalIDs["isrc"], market)
}

func isIsrcForMarket(isrc string, market string) bool {
	if market == "" {
		return true
	}
	return strings.Contains(isrc, market)
}
