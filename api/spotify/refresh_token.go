package spotify

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pp-develop/music-timer-api/model"
	"golang.org/x/oauth2"
)

func RefreshToken(user model.User) (*oauth2.Token, error) {
	var response = &oauth2.Token{}

	err := godotenv.Load()
	if err != nil {
		return response, err
	}

	values := url.Values{}
	values.Add("refresh_token", user.RefreshToken)
	values.Add("grant_type", "refresh_token")

	endopint := "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", endopint, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("SPOTIFY_ID")+":"+os.Getenv("SPOTIFY_SECRET"))))

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return response, err
	}

	body, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
