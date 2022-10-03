package internal

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func AddItemsPlaylist(playlistId string, token string) {
	endpoints := "https://api.spotify.com/v1/playlists/" + playlistId + "/tracks"
	req, _ := http.NewRequest("POST", endpoints, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	q := req.URL.Query()
	q.Add("uris", "spotify:track:4iV5W9uYEdYUVa79Axb7Rh,spotify:track:1301WleyT98MSxVHPZCA6M")
	q.Add("position", "0")
	req.URL.RawQuery = q.Encode()

	client := new(http.Client)
	resp, err := client.Do(req)
	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("%s", dumpResp)

	if err != nil {
		log.Println("httprequest error", err)
	}
	defer resp.Body.Close()
}
