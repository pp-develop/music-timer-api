package spotify

import (
	"net/http"
	"strings"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func AddItemsPlaylist(playlistId string, tracks []model.Track, token string) bool {
	endpoints := "https://api.spotify.com/v1/playlists/" + playlistId + "/tracks"
	req, _ := http.NewRequest("POST", endpoints, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	var uris []string
	for _, v := range tracks {
		uri := strings.Replace(v.Uri, "https://open.spotify.com/track/", "spotify:track:", 1)
		uris = append(uris, uri)
	}

	q := req.URL.Query()
	q.Add("uris", strings.Join(uris, ","))
	q.Add("position", "0")
	req.URL.RawQuery = q.Encode()
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	if resp.StatusCode != 201{
		return false
	}
	defer resp.Body.Close()

	return true
}
