package main

import (
	"github.com/pp-develop/music-timer-api/router"
)

func main() {
	router := router.Create()
	router.Run(":8080")
}
