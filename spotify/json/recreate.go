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

	err := createJson(db)
	if err != nil {
		log.Printf("Error creating JSON: %v", err)
		return err
	}

	return nil
}
