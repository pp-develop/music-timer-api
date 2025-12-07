package search

import (
	"database/sql"
	"log/slog"
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
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		slog.Warn("failed to bind JSON, using empty market", slog.Any("error", err))
	}

	market := strings.ToUpper(requestBody.Market)
	slog.Info("save tracks started", slog.String("market", market))

	tracks, err := spotify.SearchTracks(market)
	if err != nil {
		slog.Error("error fetching from Spotify API", slog.Any("error", err))
		return err
	}
	slog.Debug("fetched tracks from Spotify API", slog.Int("track_count", len(tracks)))

	savedCount, err := saveTracks(db, tracks, market, true)
	if err != nil {
		slog.Error("error saving to DB", slog.Any("error", err))
		return err
	}

	slog.Info("save tracks complete", slog.Int("fetched", len(tracks)), slog.Int("saved", savedCount), slog.Duration("duration", time.Since(start)))

	// 大量データ処理後にGCを実行してメモリを解放
	runtime.GC()

	return nil
}

func saveTracks(db *sql.DB, tracks []spotifylibrary.FullTrack, market string, validate bool) (int, error) {
	// URIをキーにして重複を除去 + バリデーション
	// validTracksには重複なし（seenで既出URIをスキップ）かつバリデーション通過のトラックのみ格納
	seen := make(map[string]bool)
	validTracks := make([]spotifylibrary.FullTrack, 0, len(tracks))

	for _, item := range tracks {
		uri := string(item.URI)

		// 重複チェック - 既に見たURIならスキップ
		if seen[uri] {
			continue
		}
		seen[uri] = true

		// バリデーション
		if validate && !validateTrack(item, market) {
			continue
		}

		// ここに到達するのは「重複なし + バリデーション通過」のみ
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
