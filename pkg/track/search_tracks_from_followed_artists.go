package track

import (
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func SearchTracksFromFollowedArtists(userId string) error {
	artist, err := database.GetFollowedArtists(userId)
	if err != nil {
		return err
	}

	for _, item := range artist {
		items, err := spotify.SearchTracksByArtists(item.Name)
		if err != nil {
			return err
		}
	
		err = SaveTracks(items)
		if err != nil {
			return err
		}
	
		err = NextSearchTracks(items)
		if err != nil {
			return err
		}
	}

	return nil
}
