package json

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
)

type TrackWithoutArtistsId struct {
	Uri        string `json:"uri"`
	DurationMs int    `json:"duration_ms"`
	Isrc       string `json:"isrc"`
}

const baseDirectory = "./pkg/json"
const fileNamePattern = "tracks_part_%d.json"

// ファイル存在キャッシュ（メモリ上に保持）
var (
	filesExistCache      bool
	filesExistCacheMutex sync.RWMutex
	cacheInitialized     bool
)

// ClearFilesExistCache はファイル存在キャッシュをクリアする（ReCreate時に呼び出す）
func ClearFilesExistCache() {
	filesExistCacheMutex.Lock()
	defer filesExistCacheMutex.Unlock()
	cacheInitialized = false
}

// deleteAllTrackFiles は既存のトラックファイルを全て削除する（ReCreate時に呼び出す）
func deleteAllTrackFiles() error {
	for i := 1; ; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break // ファイルが存在しない = 全て削除完了
		}
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to delete %s: %w", filePath, err)
		}
		log.Printf("Deleted file: %s", filePath)
	}
	return nil
}

// setFilesExistCache はキャッシュを設定する（createJson成功後に呼び出す）
func setFilesExistCache(exists bool) {
	filesExistCacheMutex.Lock()
	defer filesExistCacheMutex.Unlock()
	filesExistCache = exists
	cacheInitialized = true
}

func exist() (bool, error) {
	filesExistCacheMutex.RLock()
	if cacheInitialized {
		result := filesExistCache
		filesExistCacheMutex.RUnlock()
		return result, nil
	}
	filesExistCacheMutex.RUnlock()

	// 初回のみ実際にファイルをチェック
	filesExistCacheMutex.Lock()
	defer filesExistCacheMutex.Unlock()

	// ダブルチェック（別のgoroutineが先に初期化した可能性）
	if cacheInitialized {
		return filesExistCache, nil
	}

	result, err := checkFilesExist()
	if err != nil {
		// エラー時はキャッシュせず、次回再試行
		return false, err
	}

	filesExistCache = result
	cacheInitialized = true

	return result, nil
}

// checkFilesExist は実際にファイルの存在をチェックする（動的ファイル数対応）
func checkFilesExist() (bool, error) {
	// 最低1ファイルは必要
	firstFilePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, 1))
	if _, err := os.Stat(firstFilePath); os.IsNotExist(err) {
		log.Printf("First file does not exist: %s", firstFilePath)
		return false, nil
	}

	// 存在するファイルをすべてチェック
	for i := 1; ; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i))

		file, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// ファイルが存在しない = 全ファイルチェック完了
				if i == 1 {
					return false, nil // 1つもない
				}
				break // i-1個のファイルが存在
			}
			return false, err
		}

		var config struct {
			Tracks []struct{} `json:"tracks"`
		}

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		file.Close()

		if err != nil {
			log.Printf("Error decoding file %s: %v", filePath, err)
			return false, err
		}

		if len(config.Tracks) == 0 {
			log.Printf("Tracks are empty in file: %s", filePath)
			return false, nil
		}
	}
	return true, nil
}

func createJson(db *sql.DB) error {
	// 1ファイルあたりのトラック数
	// メモリ効率のため5万件（約5MB）を上限として分割
	const tracksPerFile = 50000

	// ディレクトリの存在確認
	if _, err := os.Stat(baseDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	pageNumber := 1

	for {
		// 1ページ分のみ取得（メモリ効率化）
		tracks, err := database.GetTracks(db, pageNumber, tracksPerFile)
		if err != nil {
			return err
		}
		if len(tracks) == 0 {
			break // データ終了
		}

		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, pageNumber))
		log.Printf("Creating file: %s", filePath)

		err = retry(3, 1*time.Second, func() error {
			return writeTracksToFileStreaming(filePath, tracks)
		})
		if err != nil {
			return err
		}

		log.Printf("Saved %d tracks to file: %s", len(tracks), filePath)

		// メモリ解放
		tracks = nil
		runtime.GC()

		pageNumber++
	}

	return nil
}

// writeTracksToFileStreaming はストリーミング方式でJSONを書き込む（メモリ効率改善）
func writeTracksToFileStreaming(filePath string, tracks []model.Track) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// JSONの開始
	if _, err := writer.WriteString(`{"tracks":[`); err != nil {
		return err
	}

	// 1件ずつJSONエンコードして書き込み（MarshalIndentを使わない）
	for i, track := range tracks {
		if i > 0 {
			if _, err := writer.WriteString(","); err != nil {
				return err
			}
		}
		trackData := TrackWithoutArtistsId{
			Uri:        track.Uri,
			DurationMs: track.DurationMs,
			Isrc:       track.Isrc,
		}
		data, err := json.Marshal(trackData)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}

	// JSONの終了
	if _, err := writer.WriteString("]}"); err != nil {
		return err
	}

	log.Printf("Saved %d tracks to file: %s", len(tracks), filePath)
	return writer.Flush()
}

func retry(attempts int, sleep time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(sleep)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func Create(db *sql.DB) error {
	// jsonがあれば何もしない
	exists, err := exist()
	if err != nil {
		log.Printf("Error checking existence: %v", err)
		return err
	}
	if exists {
		return nil
	}

	err = createJson(db)
	if err != nil {
		log.Printf("Error creating JSON: %v", err)
		return err
	}
	log.Println("creating JSON")

	// 作成成功後、キャッシュをtrueに設定
	setFilesExistCache(true)

	return nil
}
