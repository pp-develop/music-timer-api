package json

import (
	"encoding/json"
	"io/ioutil"
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

var filePath = "config.json"

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

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &cm.config); err != nil {
		return err
	}

	// if cm.config.Tracks == nil {
	// 	cm.config.Tracks = []model.Track{}
	// }
	// Tracksを初期化する
	cm.config.Tracks = []model.Track{}

	return nil
}

func (cm *ConfigManager) save() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	bytes, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(cm.configFilePath, bytes, 0644); err != nil {
		return err
	}
	return nil
}

func Create() error {
	configManager, err := NewConfigManager(filePath)
	if err != nil {
		return err
	}

	allTracks, err := database.GetAllTracks()
	if err != nil {
		return err
	}

	configManager.mutex.Lock()
	configManager.config.Tracks = append(configManager.config.Tracks, allTracks...)
	configManager.mutex.Unlock()

	if err := configManager.save(); err != nil {
		return err
	}
	return nil
}
