package spotify

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var config = &clientcredentials.Config{
	ClientID:     os.Getenv("SPOTIFY_ID"),
	ClientSecret: os.Getenv("SPOTIFY_SECRET"),
	TokenURL:     spotifyauth.TokenURL,
}

func getClient() (*spotify.Client, error) {
	token, err := config.Token(context.Background())
	if err != nil {
		return nil, err
	}

	clientcredentialHttpClient := spotifyauth.New().Client(context.Background(), token)
	client := spotify.New(clientcredentialHttpClient, spotify.WithRetry(true))
	return client, nil
}

func SearchTracks() (*spotify.SearchResult, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	options := []spotify.RequestOption{spotify.Market("JP"), spotify.Limit(50)}

	results, err := client.Search(context.Background(), getRandomQuery(), spotify.SearchTypeTrack, options...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func SearchTracksByArtists(artistName string) (*spotify.SearchResult, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	client, err := getClient()
	if err != nil {
		return nil, err
	}

	options := []spotify.RequestOption{spotify.Market("JP"), spotify.Limit(50)}
	results, err := client.Search(context.Background(), "artist:"+artistName, spotify.SearchTypeTrack, options...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func NextSearchTracks(items *spotify.SearchResult) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))
	err = client.NextTrackResults(ctx, items)
	if err != nil {
		return err
	}

	return nil
}

var (
	japaneseChars = []rune("あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをん")
	chars         = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// n文字のランダムな日本語の文字列を生成する
func randomJapaneseString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = japaneseChars[rand.Intn(len(japaneseChars))]
	}
	return string(b)
}

// n文字のランダムな文字列を生成する
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// 1か2の数字をランダムに作成する関数
func randomOneOrTwo() int {
	return rand.Intn(2) + 1
}

func getRandomQuery() string {
	rand.Seed(time.Now().UnixNano())

	// Getting random character
	num := "0123"
	shuffled_num := num[rand.Intn(len(num))]

	random_query := ""
	switch string(shuffled_num) {
	case "0":
		random_query = randomString(randomOneOrTwo()) + "*"
	case "1":
		random_query = "*" + randomString(randomOneOrTwo()) + "*"
	case "2":
		random_query = randomJapaneseString(randomOneOrTwo()) + "*"
	case "3":
		random_query = "*" + randomJapaneseString(randomOneOrTwo()) + "*"
	}
	return random_query
}
