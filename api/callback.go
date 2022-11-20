package api

import (
	"log"
	"io"
	"net/http"
	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
)

func Callback(c *gin.Context) bool {
	code := c.Query("code")
	state := c.Query("state")
	// api/tokenエンドポイントへhttpリクエスト
	success, response := RequestApiToken(code)
	if !success {
		return false
	}

	// TODO stateの検証
	log.Println(state)

	// userid取得
	isGet, user := GetMe(response.AccessToken)
	if !isGet {
		return false
	}

	// リフレッシュトークンをDBに保存
	success = database.SaveRefreshToken(response, user.Id)
	if !success {
		return false
	}

	// sessionにuseridを格納
    session := sessions.Default(c)
	session.Set("userId", user.Id)
	session.Save()

	return true
}

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
