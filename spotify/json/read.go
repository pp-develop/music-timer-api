package json

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/pp-develop/music-timer-api/model"
)

// TracksJSON はJSONファイルの構造体（型付きで直接デコード用）
type TracksJSON struct {
	Tracks []model.Track `json:"tracks"`
}

func GetAllTracks(db *sql.DB) ([]model.Track, error) {
	// ファイルの作成
	err := Create(db)
	if err != nil {
		return nil, err
	}

	// 存在するファイル数をカウント
	fileCount := countExistingFiles()
	if fileCount == 0 {
		return nil, fmt.Errorf("no track files found")
	}

	// ランダムにファイルを選択
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	partNumber := r.Intn(fileCount) + 1
	randomFilePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, partNumber))

	// ファイルを読み込む
	data, err := readJSONFileWithRetry(randomFilePath, 3)
	if err != nil {
		return nil, err
	}

	return data.Tracks, nil
}

// countExistingFiles は存在するトラックファイルの数を返す
func countExistingFiles() int {
	count := 0
	for i := 1; ; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i))
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		count++
	}
	return count
}

func readJSONFileWithRetry(filePath string, retries int) (*TracksJSON, error) {
	var lastErr error

	for i := 0; i < retries; i++ {
		file, openErr := os.Open(filePath)
		if openErr != nil {
			slog.Error("error opening file", slog.String("file_path", filePath), slog.Any("error", openErr))
			return nil, openErr
		}

		var data TracksJSON
		decoder := json.NewDecoder(file)
		err := decoder.Decode(&data)
		file.Close() // ループ内なのでdeferではなく即座にクローズ

		if err == nil {
			return &data, nil
		}

		lastErr = err
		slog.Warn("error decoding JSON from file", slog.String("file_path", filePath), slog.Int("attempt", i+1), slog.Int("retries", retries), slog.Any("error", err))
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("failed to read and decode JSON from file %s after %d attempts: %w", filePath, retries, lastErr)
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
