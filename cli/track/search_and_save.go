package main

import (
	"log"
	"os"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/pkg/search"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "search and save track",
		Usage: "search spotify tracks to DB save",
		Action: func(*cli.Context) error {
			log.Println("start search and save track")

			db, err := database.GetDatabaseInstance(database.CockroachDB{})
			if err != nil {
				log.Println(err)
				return err
			}

			err = search.SaveTracks(db)
			if err != nil {
				log.Println(err)
				return err
			}

			log.Println("end search and save track")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
