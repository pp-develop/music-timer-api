package main

import (
    "fmt"
    "log"
    "os"

    "github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name:  "get_track",
        Usage: "get spotify tracks to DB save",
        Action: func(*cli.Context) error {
            fmt.Println("Hello")
            return nil
        },
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }
}
