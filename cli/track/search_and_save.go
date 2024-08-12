package main

import (
	"log"
	"os"

	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/search"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "search and save track",
		Usage: "search spotify tracks to DB save",
		Action: func(*cli.Context) error {
			log.Println("start search and save track")

			err := search.SaveTracks()
			if err != nil {
				log.Println(err)
			}

			log.Println("end search and save track")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
