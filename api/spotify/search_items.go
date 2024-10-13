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

func SearchTracks(market string) ([]spotify.FullTrack, error) {
	// 環境変数をロード
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	// コンテキストとクライアント認証情報の初期化
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	// トークンを使ってSpotifyクライアントを作成
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	// 検索オプションの設定 (marketが指定されていれば追加)
	options := []spotify.RequestOption{spotify.Limit(50)}
	if market != "" {
		options = append(options, spotify.Market(market))
	}

	// トラック検索の開始
	var fullTracks []spotify.FullTrack
	query := getRandomQuery()

	// 検索とページング処理
	results, err := client.Search(ctx, query, spotify.SearchTypeTrack, options...)
	if err != nil {
		return nil, err
	}
	fullTracks = append(fullTracks, results.Tracks.Tracks...)

	// ページング処理
	for {
		var prevUrl = results.Tracks.Next

		// 次のページを取得
		err = client.NextTrackResults(ctx, results)
		if err != nil {
			return nil, err
		}
		fullTracks = append(fullTracks, results.Tracks.Tracks...)

		if prevUrl == results.Tracks.Next {
			break
		}
	}

	return fullTracks, nil
}

func SearchTracksByArtists(artistName string, market string) (*spotify.SearchResult, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, err
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient, spotify.WithRetry(true))

	// 検索オプションを設定 (market が空文字でない場合のみマーケットを指定)
	options := []spotify.RequestOption{spotify.Limit(50)}
	if market != "" {
		options = append(options, spotify.Market(market))
	}

	results, err := client.Search(context.Background(), "artist:"+artistName, spotify.SearchTypeTrack, options...)
	if err != nil {
		return nil, err
	}

	return results, nil
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
