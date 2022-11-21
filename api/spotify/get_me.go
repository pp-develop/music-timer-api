package spotify

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"

	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

func GetMe(code string) (bool, model.User) {
	var response model.User

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return false, response
	}

	endopint := "https://api.spotify.com/v1/me"
	req, _ := http.NewRequest("GET", endopint, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+code)

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
