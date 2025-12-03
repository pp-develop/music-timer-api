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

// checkFilesExist は実際にファイルの存在をチェックする
func checkFilesExist() (bool, error) {
	for i := 1; i <= 10; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i))
		log.Printf("Checking file: %s", filePath)

		file, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("File does not exist: %s", filePath)
				return false, nil
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
	allTracks, err := database.GetAllTracks(db)
	if err != nil {
		return err
	}

	// トラックを10個のファイルに分割
	numFiles := 10
	numTracksPerFile := (len(allTracks) + numFiles - 1) / numFiles

	// ディレクトリの存在確認
	if _, err := os.Stat(baseDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(baseDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	for i := 0; i < numFiles; i++ {
		start := i * numTracksPerFile
		end := start + numTracksPerFile
		if end > len(allTracks) {
			end = len(allTracks)
		}

		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i+1))
		log.Printf("Creating file: %s", filePath)

		err := retry(3, 1*time.Second, func() error {
			return writeTracksToFileStreaming(filePath, allTracks[start:end])
		})
		if err != nil {
			return err
		}

		// 各ファイル書き込み後にGCを実行してメモリを解放
		runtime.GC()
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
