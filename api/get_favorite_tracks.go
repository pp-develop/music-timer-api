package api

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetFavoriteTracks(specify_ms int, userId string) ([]model.Track, error) {
	var tracks []model.Track

	c1 := make(chan []model.Track, 1)
	go func() {
		successGetTracks := false
		for !successGetTracks {
			allTracks, _ := getAllTracksByFavoriteArtists(userId)
			successGetTracks, tracks = getFavoriteTracksBySpecifyTime(allTracks, specify_ms)
		}
		c1 <- tracks
	}()

	select {
	case <-c1:
		return tracks, nil
	case <-time.After(30 * time.Second):
		return tracks, errors.New("get tracks: time out")
	}
}

func getAllTracksByFavoriteArtists(userId string) ([]model.Track, error) {
	var tracks []model.Track

	favoriteArtists, err := database.GetFavoriteAllArtists(userId)
	if err != nil {
		log.Println(err)
		return tracks, err
	}

	var artistName string
	for _, v := range favoriteArtists {
		v.Name = strings.Replace(v.Name, "'", "", -1)
		artistName += "artists_name like " + "'%" + v.Name + "%' OR "
	}
	artistName = artistName[0 : len(artistName)-3]

	favoriteTracks, err := database.GetAllTracksByArtistsName(artistName)
	if err != nil {
		log.Println(err)
		return tracks, err
	}
	return favoriteTracks, nil
}

func getFavoriteTracksBySpecifyTime(allTracks []model.Track, specify_ms int) (bool, []model.Track) {
	var tracks []model.Track
	var sum_ms int

	// tracksの合計分数が指定された分数を超過したらループを停止
	for _, v := range allTracks {
		tracks = append(tracks, v)
		sum_ms += v.DurationMs
		if sum_ms > specify_ms {
			break
		}
	}

	// tracksから要素を1つ削除
	tracks = tracks[:len(tracks)-1]

	// 指定分数とtracksの合計分数の差分を求める
	sum_ms = 0
	var diff_ms int
	for _, v := range tracks {
		sum_ms += v.DurationMs
	}
	diff_ms = specify_ms - sum_ms

	// 誤差が15秒以内は許容
	allowance_ms := 15000
	if diff_ms == allowance_ms {
		return true, tracks
	}

	// 差分を埋めるtrackを取得
	var isGetTrack bool
	getTrack := database.GetTrackByMsec(diff_ms)
	if len(getTrack) > 0 {
		isGetTrack = true
		tracks = append(tracks, getTrack...)
	}
	return isGetTrack, tracks
}
