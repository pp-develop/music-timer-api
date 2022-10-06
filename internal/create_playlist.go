package internal

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type RequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Response struct {
	Id string `json:"id"`
}

func CreatePlaylist(userid string, ms int, token string) (bool, string) {
	requestBody := &RequestBody{
		Name:        strconv.Itoa(ms/60000) + "分",
		Description: "",
	}
	jsonString, _ := json.Marshal(requestBody)

	endopint := "https://api.spotify.com/v1/users/" + userid + "/playlists"
	req, _ := http.NewRequest("POST", endopint, bytes.NewBuffer(jsonString))
	req.Header.Set("Authorization", "Bearer "+token)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return false, ""
	}

	body, _ := io.ReadAll(resp.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, ""
	}
	return true, response.Id
}
