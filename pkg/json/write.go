package json

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
)

type Config struct {
	Foo string   `json:"foo"`
	Bar int      `json:"bar"`
	Baz []string `json:"baz"`
}

type ConfigManager struct {
	configFilePath string
	config         Config
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

	// Bazがなければ初期化する
	if cm.config.Baz == nil {
		cm.config.Baz = []string{}
	}

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

func Write() {
	configManager, err := NewConfigManager(filePath)
	if err != nil {
		panic(err)
	}

	configManager.mutex.Lock()
	configManager.config.Baz = append(configManager.config.Baz, "new_baz")
	configManager.mutex.Unlock()

	if err := configManager.save(); err != nil {
		panic(err)
	}
}
