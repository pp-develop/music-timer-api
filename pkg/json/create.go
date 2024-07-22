package json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

type Json struct {
	Tracks []model.Track `json:"tracks"`
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

	file, err := os.Open(cm.configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cm.config); err != nil {
		return err
	}

	// if cm.config.Tracks == nil {
	// 	cm.config.Tracks = []model.Track{}
	// }
	// Tracksを初期化する
	cm.config.Tracks = []model.Track{}

	return nil
}

func exist() (bool, error) {
	for i := 1; i <= 10; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i+1))
		file, err := os.Open(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				// ファイルが存在しない場合
				return false, nil
			}
			// その他のエラー
			return false, err
		}

		var config struct {
			Tracks []struct{} `json:"tracks"`
		}

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&config); err != nil {
			file.Close() // デコードエラーが発生した場合、ファイルを閉じます
			return false, err
		}
		file.Close() // デコードが完了したらファイルを閉じます

		if len(config.Tracks) == 0 {
			// トラックが空の場合
			return false, nil
		}
	}
	// すべてのファイルが存在し、空でない
	return true, nil
}

func createJson() error {
	allTracks, err := database.GetAllTracks()
	if err != nil {
		return err
	}

	// トラックを10個の部分に分割
	numFiles := 10
	numTracksPerFile := (len(allTracks) + numFiles - 1) / numFiles // 均等に分割できない場合は切り上げ

	for i := 0; i < numFiles; i++ {
		start := i * numTracksPerFile
		end := start + numTracksPerFile
		if end > len(allTracks) {
			end = len(allTracks)
		}

		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i+1))
		configManager, err := NewConfigManager(filePath)
		if err != nil {
			return err
		}

		err = configManager.saveToFile(filePath, allTracks[start:end])
		if err != nil {
			return err
		}
	}

	return nil
}

func (cm *ConfigManager) saveToFile(filePath string, tracks []model.Track) error {
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

func Create() error {
	// jsonがあれば何もしない
	exists, err := exist()
	if err != nil {
		log.Printf("Error checking existence: %v", err)
		return err
	}
	if exists {
		return nil
	}

	err = createJson()
	if err != nil {
		log.Printf("Error creating JSON: %v", err)
		return err
	}

	return nil
}
