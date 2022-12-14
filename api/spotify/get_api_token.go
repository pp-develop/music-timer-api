package spotify

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

// https://developer.spotify.com/documentation/general/guides/authorization/code-flow/
func GetApiTokenForAuthzCode(code string) (bool, model.ApiTokenResponse) {
	var response model.ApiTokenResponse

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, response
	}

	values := url.Values{}
	values.Add("code", code)
	values.Add("SPOTIFY_REDIRECT_URI", os.Getenv("SPOTIFY_REDIRECT_URI"))
	values.Add("grant_type", "authorization_code")

	endopint := "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", endopint, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("SPOTIFY_ID")+":"+os.Getenv("SPOTIFY_SECRET"))))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return false, response
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Print(err)
		return false, response
	}

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Print(err)
		return false, response
	}
	return true, response
}
