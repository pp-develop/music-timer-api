package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type RequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreatePlaylist(userid string, token string) {
	requestBody := &RequestBody{
		Name:        "pp!!!",
		Description: "test",
	}
	jsonString, _ := json.Marshal(requestBody)
	
	endopint := "https://api.spotify.com/v1/users/" + userid + "/playlists"
	req, _ := http.NewRequest("POST", endopint, bytes.NewBuffer(jsonString))
	req.Header.Set("Authorization", "Bearer "+token)
	client := new(http.Client)
	resp, err := client.Do(req)
	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("%s", dumpResp)

	if err != nil {
		log.Println("httprequest error", err)
	}
	defer resp.Body.Close()
}
