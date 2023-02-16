package api

import (
	"log"
	"strings"
	"time"

	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetFollowedArtistsTracksBySpecifyTime(specify_ms int, userId string) ([]model.Track, error) {
	var tracks []model.Track

	c1 := make(chan []model.Track, 1)
	go func() {
		success := false
		for !success {
			tracks, _ = GetFollowedArtistsTracks(userId)
			success, tracks = MakeTracksBySpecifyTime(tracks, specify_ms)
		}
		c1 <- tracks
	}()

	select {
	case <-c1:
		return tracks, nil
	case <-time.After(30 * time.Second):
		return nil, model.ErrTimeoutCreatePlaylist
	}
}

func GetFollowedArtistsTracks(userId string) ([]model.Track, error) {
	var tracks []model.Track

	followedArtists, err := database.GetFollowedArtists(userId)
	if err != nil {
		log.Println(err)
		return tracks, err
	}

	var artistName string
	for _, v := range followedArtists {
		v.Name = strings.Replace(v.Name, "'", "", -1)
		artistName += "artists_name like " + "'%" + v.Name + "%' OR "
	}
	artistName = artistName[0 : len(artistName)-3]

	favoriteTracks, err := database.GetTracksByArtistsName(artistName)
	if err != nil {
		log.Println(err)
		return tracks, err
	}
	return favoriteTracks, nil
}
