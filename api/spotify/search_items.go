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

// func __SearchTracks(code string) (bool, model.SearchTracksResponse) {
// 	var response model.SearchTracksResponse

// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Println("Error loading .env file")
// 	}

// 	endpoint := "https://api.spotify.com/v1/search"
// 	req, _ := http.NewRequest("GET", endpoint, nil)
// 	req.Header.Set("Accept", "application/json")
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+code)
// 	q := req.URL.Query()
// 	q.Add("q", "35.7114441")
// 	q.Add("type", "track")
// 	req.URL.RawQuery = q.Encode()

// 	client := new(http.Client)
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Print(err)
// 		return false, response
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != 200 {
// 		log.Print(err)
// 		return false, response
// 	}

// 	body, _ := io.ReadAll(resp.Body)
// 	err = json.Unmarshal(body, &response)
// 	if err != nil {
// 		log.Print(err)
// 		return false, response
// 	}
// 	return true, response
// }
