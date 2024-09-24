package json

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

func GetAllTracks(db *sql.DB) ([]model.Track, error) {
	// ファイルの作成
	err := Create(db)
	if err != nil {
		return nil, err
	}

	// ランダムにファイルを選択するための設定
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	// ランダムな数（1から10）を生成
	partNumber := r.Intn(10) + 1
	randomFilePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, partNumber))

	// 非同期でファイルを読み込む
	tracksChan := make(chan []model.Track)
	errChan := make(chan error)

	go func() {
		defer close(errChan) // エラーチャネルをクローズ
		data, err := readJSONFileWithRetry(randomFilePath, 3)
		if err != nil {
			errChan <- err
			return
		}

		// 配列型のキーを取得
		key := "tracks"
		value, ok := data[key]
		if !ok {
			errChan <- fmt.Errorf("key %s not found", key)
			return
		}

		// 配列にキャスト
		array, ok := value.([]interface{})
		if !ok {
			errChan <- fmt.Errorf("key %s is not an array", key)
			return
		}

		// []interface{}型の配列を[]Track型に変換
		tracks := make([]model.Track, 0, len(array))
		for _, v := range array {
			item := v.(map[string]interface{})
			track := model.Track{
				Uri:        item["uri"].(string),
				DurationMs: int(item["duration_ms"].(float64)),
				Isrc:       item["isrc"].(string),
			}
			tracks = append(tracks, track)
		}
		tracksChan <- tracks
		close(tracksChan) // トラックチャネルをクローズ
	}()

	select {
	case tracks := <-tracksChan:
		return tracks, nil
	case err := <-errChan:
		return nil, err
	}
}

func readJSONFileWithRetry(filePath string, retries int) (map[string]interface{}, error) {
	var data map[string]interface{}
	var err error

	for i := 0; i < retries; i++ {
		file, openErr := os.Open(filePath)
		if openErr != nil {
			log.Printf("Error opening file %s: %v", filePath, openErr)
			return nil, openErr
		}
		defer file.Close() // ファイルを関数終了時に自動的に閉じる

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&data)
		if err == nil {
			// 成功した場合
			return data, nil
		}

		// エラーが発生した場合、リトライの前にログを出力
		log.Printf("Error decoding JSON from file %s (attempt %d/%d): %v", filePath, i+1, retries, err)
		time.Sleep(1 * time.Second) // 少し待ってから再試行
	}
	return nil, fmt.Errorf("failed to read and decode JSON from file %s after %d attempts: %w", filePath, retries, err)
}

func ShuffleTracks(tracks []model.Track) []model.Track {
	// Fisher-Yates アルゴリズムを使って、スライスの要素をランダムに並び替える
	n := len(tracks)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		tracks[i], tracks[j] = tracks[j], tracks[i]
	}
	return tracks
}

func GetTrackByMsec(allTracks []model.Track, msec int) []model.Track {
	tracks := []model.Track{}
	for _, track := range allTracks {
		if track.DurationMs == msec {
			tracks = append(tracks, track)
			break
		}
	}
	return tracks
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
