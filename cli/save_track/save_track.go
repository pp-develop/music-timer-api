package main

import (
	"log"
	"os"

	"github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "save_track",
		Usage: "search spotify tracks to DB save",
		Action: func(*cli.Context) error {
			log.Println("start save track")

			err := track.SearchTracks()
			if err != nil {
				log.Println(err)
			}

			log.Println("end save track")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
