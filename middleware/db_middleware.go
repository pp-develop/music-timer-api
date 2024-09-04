package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/database"
)

func DBMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db, err := database.GetDatabaseInstance(database.CockroachDB{})
		if err != nil {
			c.JSON(500, gin.H{"error": "Database connection error"})
			c.Abort()
			return
		}
		c.Set("db", db)
		c.Next()
	}
}
