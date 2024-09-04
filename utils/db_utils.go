package utils

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func GetDB(c *gin.Context) (*sql.DB, bool) {
	db, exists := c.Get("db")
	if !exists {
		return nil, false
	}

	dbInstance, ok := db.(*sql.DB)
	if !ok {
		return nil, false
	}
	return dbInstance, true
}
