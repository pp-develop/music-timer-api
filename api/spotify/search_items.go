package spotify

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

// getMemStats はメモリ統計を文字列で返す
func getMemStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Alloc=%dMB, Sys=%dMB, NumGC=%d",
		m.Alloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)
}

func SearchTracks(market string) ([]spotify.FullTrack, error) {
	start := time.Now()
	query := getRandomQuery()

	slog.Info("search tracks started",
		slog.String("query", query),
		slog.String("market", market))

	// コンテキストとクライアント認証情報の初期化
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_ID"),
		ClientSecret: os.Getenv("SPOTIFY_SECRET"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		slog.Error("token acquisition failed",
			slog.String("query", query),
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err))
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

	// 検索とページング処理
	results, err := client.Search(ctx, query, spotify.SearchTypeTrack, options...)
	if err != nil {
		slog.Error("initial search failed",
			slog.String("query", query),
			slog.String("market", market),
			slog.Duration("duration", time.Since(start)),
			slog.Any("error", err))
		return nil, err
	}
	fullTracks = append(fullTracks, results.Tracks.Tracks...)
	pageCount := 1

	// ページング処理
	// メモリ効率のため200ページ（10,000件）を上限とする
	const maxPages = 200

	for results.Tracks.Next != "" && pageCount < maxPages {
		pageCount++

		// 次のページを取得
		err = client.NextTrackResults(ctx, results)
		if err != nil {
			slog.Error("paging failed",
				slog.String("query", query),
				slog.String("market", market),
				slog.Int("page", pageCount),
				slog.Int("fetched", len(fullTracks)),
				slog.Duration("duration", time.Since(start)),
				slog.Any("error", err))
			return nil, err
		}
		fullTracks = append(fullTracks, results.Tracks.Tracks...)

		// 進捗ログ（20ページごと = 約1000件ごと）
		if pageCount%20 == 0 {
			slog.Info("search tracks progress",
				slog.String("query", query),
				slog.Int("page", pageCount),
				slog.Int("fetched", len(fullTracks)),
				slog.Duration("duration", time.Since(start)),
				slog.String("memory", getMemStats()))
		}
	}

	// 上限に達した場合はログ出力
	if pageCount >= maxPages {
		slog.Warn("reached max pages limit",
			slog.String("query", query),
			slog.Int("max_pages", maxPages))
	}

	slog.Info("search tracks completed",
		slog.String("query", query),
		slog.String("market", market),
		slog.Int("pages", pageCount),
		slog.Int("tracks", len(fullTracks)),
		slog.Duration("duration", time.Since(start)))

	return fullTracks, nil
}

func SearchTracksByArtists(artistName string, market string) (*spotify.SearchResult, error) {
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
