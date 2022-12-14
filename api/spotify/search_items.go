package spotify

import (
	"context"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2/clientcredentials"
)

func SearchTracks() *spotify.SearchResult {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	options := []spotify.RequestOption{spotify.Market("JP")}
	results, err := client.Search(ctx, getRandomQuery(), spotify.SearchTypeTrack, options...)
	if err != nil {
		log.Fatal(err)
	}

	return results
}

func getRandomQuery() string {
	rand.Seed(time.Now().UnixNano())

	str := "abcdefghijklmnopqrstuvwxyz"
	num := "01"

	// Getting random character
	shuffled_str := str[rand.Intn(len(str))]
	shuffled_num := num[rand.Intn(len(num))]

	random_query := ""
	switch string(shuffled_num) {
	case "0":
		random_query = string(shuffled_str) + "%"
	case "1":
		random_query = "%" + string(shuffled_str) + "%"
	}
	return random_query
}

func NextSearchTracks(items *spotify.SearchResult) (bool, *spotify.SearchResult) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	err = client.NextTrackResults(ctx, items)
	if err != nil {
		log.Print(err)
		return false, nil
	}

	return true, items
}
