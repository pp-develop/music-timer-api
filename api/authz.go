package api

import (
	"encoding/base64"
	"encoding/json"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func Authz() (bool, string) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, ""
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(os.Getenv("REDIRECT_URI")),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistModifyPublic),
		spotifyauth.WithClientID(os.Getenv("CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("CLIENT_SECRET")),
	)
	state := uuid.New()
	url := auth.AuthURL(state.String())
	return true, url
}

func RequestApiToken(code string) (bool, model.ApiTokenResponse) {
	var response model.ApiTokenResponse

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, response
	}

	values := url.Values{}
	values.Add("code", code)
	values.Add("redirect_uri", os.Getenv("REDIRECT_URI"))
	values.Add("grant_type", "authorization_code")

	endopint := "https://accounts.spotify.com/api/token"
	req, _ := http.NewRequest("POST", endopint, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID")+":"+os.Getenv("CLIENT_SECRET"))))

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
