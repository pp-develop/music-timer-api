package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
)

type Track struct {
	Id          string `json:"id"`
	Description int    `json:"description"`
}

func Read(key string) ([]Track, error) {
	// ファイルの読み込み
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// JSONのパース
	var data map[string]interface{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}

	// 配列型のキーを取得
	value, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("Key %s not found", key)
	}

	// 配列にキャスト
	array, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Key %s is not an array", key)
	}

	// []interface{}型の配列を[]Track型に変換
	tracks := make([]Track, 0, len(array))
	for _, v := range array {
		item := v.(map[string]interface{})
		track := Track{
			Id:          item["id"].(string),
			Description: int(item["description"].(float64)), // float64型をint型に変換
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func ShuffleTracks(tracks []Track) []Track {
	// Fisher-Yates アルゴリズムを使って、スライスの要素をランダムに並び替える
	n := len(tracks)
	for i := n - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		tracks[i], tracks[j] = tracks[j], tracks[i]
	}
	return tracks
}
