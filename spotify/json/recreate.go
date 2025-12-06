package json

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

func ReCreate(c *gin.Context) error {
	start := time.Now()
	log.Printf("[ReCreate] Start - %s", getMemStats())

	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// キャッシュをクリア
	ClearFilesExistCache()

	// 既存ファイルを削除（ファイル数が減った場合に古いファイルが残るのを防ぐ）
	if err := deleteAllTrackFiles(); err != nil {
		log.Printf("[ReCreate] Error deleting old files: %v", err)
		return err
	}
	log.Printf("[ReCreate] Deleted old files - %s", getMemStats())

	// 新規作成
	err := createJson(db)
	if err != nil {
		log.Printf("[ReCreate] Error creating JSON: %v", err)
		return err
	}

	log.Printf("[ReCreate] Complete - duration=%v, %s", time.Since(start), getMemStats())
	return nil
}
