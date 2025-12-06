package json

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
)

func ReCreate(c *gin.Context) error {
	db, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// キャッシュをクリア
	ClearFilesExistCache()

	// 既存ファイルを削除（ファイル数が減った場合に古いファイルが残るのを防ぐ）
	if err := deleteAllTrackFiles(); err != nil {
		log.Printf("Error deleting old files: %v", err)
		return err
	}

	// 新規作成
	err := createJson(db)
	if err != nil {
		log.Printf("Error creating JSON: %v", err)
		return err
	}

	return nil
}
