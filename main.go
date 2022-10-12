package main

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/router"
)

func main() {
	router := router.Create()
	router.Run(":8080")
}
