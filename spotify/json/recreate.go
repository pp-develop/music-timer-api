package json

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

func ReCreate(c *gin.Context) error {
	start := time.Now()
	slog.Info("recreate started", slog.String("mem_stats", getMemStats()))

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// キャッシュをクリア
	ClearFilesExistCache()

	// 既存ファイルを削除（ファイル数が減った場合に古いファイルが残るのを防ぐ）
	if err := deleteAllTrackFiles(); err != nil {
		slog.Error("error deleting old files", slog.Any("error", err))
		return err
	}
	slog.Debug("deleted old files", slog.String("mem_stats", getMemStats()))

	// 新規作成
	err := createJson(db)
	if err != nil {
		slog.Error("error creating JSON", slog.Any("error", err))
		return err
	}

	slog.Info("recreate complete", slog.Duration("duration", time.Since(start)), slog.String("mem_stats", getMemStats()))
	return nil
}
