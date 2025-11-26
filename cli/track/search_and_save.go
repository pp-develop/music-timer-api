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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "market",
				Aliases: []string{"m"},
				Usage:   "market code (e.g., JP, US)",
				Value:   "",
			},
		},
		Action: func(c *cli.Context) error {
			log.Println("start search and save track")

			db, err := database.GetDatabaseInstance(database.CockroachDB{})
			if err != nil {
				log.Println(err)
				return err
			}

			market := c.String("market")
			err = search.SaveTracksForCLI(db, market)
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
