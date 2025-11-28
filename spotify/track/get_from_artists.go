package track

import (
	"context"
	"database/sql"
	"log"
	"time"

	commontrack "github.com/pp-develop/music-timer-api/pkg/common/track"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/spotify/json"
)

func GetTracksFromArtists(db *sql.DB, specify_ms int, artistIds []string, userId string) ([]model.Track, error) {
	// Phase 1: データ取得と検証（即座にエラー判定）
	var artists []model.Artists
	for _, id := range artistIds {
		artists = append(artists, model.Artists{Id: id})
	}

	followedArtistsTracks, err := getSpecifyArtistsAllTracks(db, artists)
	if err != nil {
		return nil, err // ErrNotFoundTracksも含む
	}

	// Phase 2: 組み合わせ計算（時間がかかる可能性がある処理）
	var tracks []model.Track
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	c1 := make(chan []model.Track, 1)
	errChan := make(chan error, 1)
	tryCountChan := make(chan int, 1) // 試行回数を送信するチャネル
	tryCount := 0                     // 試行回数をカウントする変数

	go func() {
		defer close(c1)
		defer close(errChan)
		defer func() {
			tryCountChan <- tryCount // goroutine終了時に試行回数を送信
			close(tryCountChan)
		}()

		success := false
		for !success {
			select {
			case <-ctx.Done(): // タイムアウトまたはキャンセル時にループを終了
				errChan <- ctx.Err()
				return
			default:
				tryCount++
				shuffleTracks := json.ShuffleTracks(followedArtistsTracks)
				success, tracks  = commontrack.MakeTracks(shuffleTracks, specify_ms)
			}
		}
		c1 <- tracks
	}()

	select {
	case tracks := <-c1:
		finalTryCount := <-tryCountChan
		log.Printf("試行回数: %d\n", finalTryCount)
		return tracks, nil
	case err := <-errChan:
		finalTryCount := <-tryCountChan
		log.Printf("タイムアウト: 試行回数: %d\n", finalTryCount)
		return nil, err
	case <-ctx.Done(): // タイムアウト時
		finalTryCount := <-tryCountChan

		// トラックの総再生時間を計算
		totalAvailableDuration := 0
		for _, track := range followedArtistsTracks {
			totalAvailableDuration += track.DurationMs
		}

		hasEnoughDuration := totalAvailableDuration >= specify_ms

		if !hasEnoughDuration {
			log.Printf("[タイムアウト] GetTracksFromArtists: トラック不足 - 必要=%d分, 利用可能=%d分, トラック数=%d, 試行回数=%d, アーティスト数=%d",
				specify_ms/commontrack.MillisecondsPerMinute,
				totalAvailableDuration/commontrack.MillisecondsPerMinute,
				len(followedArtistsTracks),
				finalTryCount,
				len(artistIds),
			)
			return nil, model.ErrNotEnoughTracks
		} else {
			log.Printf("[タイムアウト] GetTracksFromArtists: 組み合わせ未発見 - 再生時間=%d分, トラック数=%d, 総再生時間=%d分, 試行回数=%d, アーティスト数=%d",
				specify_ms/commontrack.MillisecondsPerMinute,
				len(followedArtistsTracks),
				totalAvailableDuration/commontrack.MillisecondsPerMinute,
				finalTryCount,
				len(artistIds),
			)
			return nil, model.ErrTimeoutCreatePlaylist
		}
	}
}

func getSpecifyArtistsAllTracks(db *sql.DB, artists []model.Artists) ([]model.Track, error) {
	var tracks []model.Track

	artistsIds := ConvertArtistsToIDs(artists)
	tracks, err := database.GetTracksByArtistIds(db, artistsIds)
	if err != nil {
		return nil, err
	}
	if len(tracks) == 0 {
		return nil, model.ErrNotFoundTracks
	}
	return tracks, nil
}

// Artists の各要素から ID フィールドを抽出します。
func ConvertArtistsToIDs(artists []model.Artists) []string {
	artistIDs := make([]string, len(artists)) // アーティストの数だけ string スライスを作成

	for i, artist := range artists {
		artistIDs[i] = artist.Id // 各アーティストの ID を抽出
	}

	return artistIDs
}
