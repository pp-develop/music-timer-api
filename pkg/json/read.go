package json

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetAllTracks() ([]model.Track, error) {
	// ファイルの作成
	err := Create()
	if err != nil {
		return nil, err
	}

	// ランダムにファイルを選択するための設定
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	// ランダムな数（1から10）を生成
	partNumber := r.Intn(10) + 1
	randomFilePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, partNumber))

	// ファイルをオープン
	file, err := os.Open(randomFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// JSONデコーダの作成
	decoder := json.NewDecoder(file)

	// JSONのパース
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	// 配列型のキーを取得
	key := "tracks"
	value, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	// 配列にキャスト
	array, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s is not an array", key)
	}

	// []interface{}型の配列を[]Track型に変換
	tracks := make([]model.Track, 0, len(array))
	for _, v := range array {
		item := v.(map[string]interface{})

		var artistsNames []string
		for _, name := range item["artists_name"].([]interface{}) {
			artistsNames = append(artistsNames, name.(string))
		}

		track := model.Track{
			Uri:         item["uri"].(string),
			DurationMs:  int(item["duration_ms"].(float64)), // float64型をint型に変換
			ArtistsName: artistsNames,
		}
		tracks = append(tracks, track)
	}
	return tracks, nil
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

func GetTrackByMsec(allTracks []model.Track, msec int) ([]model.Track, error) {
	tracks := []model.Track{}
	for _, track := range allTracks {
		if track.DurationMs == msec {
			tracks = append(tracks, track)
			break
		}
	}
	return tracks, nil
}

func GetTracksByArtistsFromAllFiles(artists []model.Artists) ([]model.Track, error) {
	// ファイルの作成
	err := Create()
	if err != nil {
		return nil, err
	}

	var allTracks []model.Track
	for i := 1; i <= 10; i++ {
		filePath := fmt.Sprintf("%s/%s", baseDirectory, fmt.Sprintf(fileNamePattern, i))
		tracks, err := getTracksByArtistsFromFile(filePath, artists)
		if err != nil {
			return nil, err
		}
		allTracks = append(allTracks, tracks...)
	}
	return allTracks, nil
}

func getTracksByArtistsFromFile(filePath string, artists []model.Artists) ([]model.Track, error) {
	// ファイルをオープン
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// JSONデコーダの作成
	decoder := json.NewDecoder(file)

	// JSONのパース
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}

	// 配列型のキーを取得
	key := "tracks"
	value, ok := data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	// 配列にキャスト
	array, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s is not an array", key)
	}

	// []interface{}型の配列を[]Track型に変換
	var tracks []model.Track
	for _, v := range array {
		item := v.(map[string]interface{})

		var artistIds []string
		for _, id := range item["artists_id"].([]interface{}) {
			artistIds = append(artistIds, id.(string))
		}

		track := model.Track{
			Uri:        item["uri"].(string),
			DurationMs: int(item["duration_ms"].(float64)), // float64型をint型に変換
			ArtistsId:  artistIds,
		}

		// アーティスト名によるフィルタリング
		for _, artist := range artists {
			if contains(artistIds, artist.Id) {
				tracks = append(tracks, track)
				break
			}
		}
	}
	return tracks, nil
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
