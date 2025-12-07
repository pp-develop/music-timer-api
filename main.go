package main

import (
	_ "github.com/pp-develop/music-timer-api/pkg/logger" // JSON形式のslogロガーを初期化
	"github.com/pp-develop/music-timer-api/router"
)

func main() {
	router := router.Create()
	router.Run(":8080")
}
