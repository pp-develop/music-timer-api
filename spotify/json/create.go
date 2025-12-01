package json

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/pp-develop/music-timer-api/database"
)

type TrackWithoutArtistsId struct {
	Uri        string `json:"uri"`
	DurationMs int    `json:"duration_ms"`
	Isrc       string `json:"isrc"`
}

type Json struct {
	Tracks []TrackWithoutArtistsId `json:"tracks"`
}

type ConfigManager struct {
	configFilePath string
	config         Json
	mutex          sync.Mutex
}

const baseDirectory = "./pkg/json"
const fileNamePattern = "tracks_part_%d.json"

func NewConfigManager(configFilePath string) (*ConfigManager, error) {
	cm := &ConfigManager{configFilePath: configFilePath}
	if err := cm.load(); err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ConfigManager) load() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// リトライを追加
	err := retry(3, 1*time.Second, func() error {
		file, err := os.Open(cm.configFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		// ファイルサイズをチェックし、空でないことを確認
		stat, err := file.Stat()
		if err != nil {
			return err
		}
		if stat.Size() == 0 {
			return fmt.Errorf("file is empty: %s", cm.configFilePath)
		}

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&cm.config); err != nil {
			return fmt.Errorf("error decoding JSON from file %s: %w", cm.configFilePath, err)
		}

		cm.config.Tracks = []TrackWithoutArtistsId{}
		return nil
	})

	return err
}

func exist() (bool, error) {
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
		file.Close() // ループ内なのでdeferではなく即座にクローズ

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
			// ArtistsIdを含まないトラックのスライスを作成
			tracksWithoutArtistsId := make([]TrackWithoutArtistsId, len(allTracks[start:end]))
			for j, track := range allTracks[start:end] {
				tracksWithoutArtistsId[j] = TrackWithoutArtistsId{
					Uri:        track.Uri,
					DurationMs: track.DurationMs,
					Isrc:       track.Isrc,
				}
			}

			configManager, err := NewConfigManager(filePath)
			if err != nil {
				return err
			}

			log.Printf("Saving tracks to file: %s", filePath)
			return configManager.saveToFile(filePath, tracksWithoutArtistsId)
		})
		if err != nil {
			return err
		}
	}

	return nil
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

func (cm *ConfigManager) saveToFile(filePath string, tracks []TrackWithoutArtistsId) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	config := Json{Tracks: tracks}
	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
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

	return nil
}
