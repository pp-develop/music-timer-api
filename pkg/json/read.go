package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetAllTracks() ([]model.Track, error) {
	// ファイルの作成
	err := Create()
	if err != nil {
		return nil, err
	}

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
	key := "tracks"
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
	tracks := make([]model.Track, 0, len(array))
	for _, v := range array {
		item := v.(map[string]interface{})
		track := model.Track{
			Uri:         item["uri"].(string),
			DurationMs:  int(item["duration_ms"].(float64)), // float64型をint型に変換
			ArtistsName: item["artists_name"].(string),
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func GetFollowedArtistsAllTracks(artists []model.Artists) ([]model.Track, error) {
	allTrack, err := GetAllTracks()
	if err != nil {
		return nil, err
	}

	tracks := filterTracksByArtist(allTrack, artists)
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

func filterTracksByArtist(tracks []model.Track, artists []model.Artists) []model.Track {
	filteredTracks := []model.Track{}
	for _, track := range tracks {
		for _, artist := range artists {
			if strings.Contains(track.ArtistsName, artist.Name) {
				filteredTracks = append(filteredTracks, track)
				break
			}
		}
	}
	return filteredTracks
}