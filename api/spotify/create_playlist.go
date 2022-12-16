package spotify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

// 1minute = 60000ms
const ONEMINUTE_TO_MSEC = 60000

func CreatePlaylist(userid string, ms int, token string) (model.CreatePlaylistResponse, error) {
	requestBody := &model.CreatePlaylistReqBody{
		Name:        strconv.Itoa(ms/ONEMINUTE_TO_MSEC) + "分",
		Description: "",
	}
	jsonString, _ := json.Marshal(requestBody)
	var response model.CreatePlaylistResponse

	endopint := "https://api.spotify.com/v1/users/" + userid + "/playlists"
	req, _ := http.NewRequest("POST", endopint, bytes.NewBuffer(jsonString))
	req.Header.Set("Authorization", "Bearer "+token)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return response, err
	}

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
