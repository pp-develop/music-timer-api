package auth

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"
	spotifyapi "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

// SpotifyCallback Webアプリとネイティブアプリの両方に対応したSpotify OAuthコールバックを処理する
func SpotifyCallback(c *gin.Context) error {
	code := c.Query("code")
	qState := c.Query("state")
	isNativeApp := isNativeApplication(c)

	log.Println("State:", qState, "IsNativeApp:", isNativeApp)

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	// Webアプリの場合はstate検証を実行
	if err := validateState(c, qState, isNativeApp); err != nil {
		return err
	}

	// 認証コードをトークンに交換
	token, err := exchangeCodeForToken(code, isNativeApp)
	if err != nil {
		return err
	}

	// ユーザー情報を取得してデータベースに保存
	user, err := getUserAndSaveToken(dbInstance, token)
	if err != nil {
		return err
	}

	// アプリタイプに応じて認証データを設定
	return setAuthenticationData(c, user.ID, isNativeApp)
}

// isNativeApplication ネイティブアプリからのリクエストかどうかを判定する
func isNativeApplication(c *gin.Context) bool {
	isNativeApp, _ := c.Get("isNativeApp")
	return isNativeApp == true
}

// validateState Webアプリ用のOAuth stateパラメータを検証する
func validateState(c *gin.Context, qState string, isNativeApp bool) error {
	// ネイティブアプリの場合はstate検証をスキップ
	if isNativeApp {
		return nil
	}

	session := sessions.Default(c)
	v := session.Get("state")
	if v == nil {
		return model.ErrFailedGetSession
	}

	state := v.(string)
	log.Println("Session state:", state)

	if state != qState {
		return model.ErrInvalidState
	}

	return nil
}

// getRedirectURI アプリタイプに応じた適切なリダイレクトURIを返す
func getRedirectURI(isNativeApp bool) string {
	if isNativeApp {
		return os.Getenv("SPOTIFY_REDIRECT_URI_NATIVE")
	}
	return os.Getenv("SPOTIFY_REDIRECT_URI")
}

// exchangeCodeForToken 認証コードをアクセストークンに交換する
func exchangeCodeForToken(code string, isNativeApp bool) (*oauth2.Token, error) {
	redirectURI := getRedirectURI(isNativeApp)
	return spotify.ExchangeSpotifyCode(code, redirectURI)
}

// getUserAndSaveToken ユーザー情報を取得してトークンをデータベースに保存する
func getUserAndSaveToken(dbInstance interface{}, token *oauth2.Token) (*spotifyapi.PrivateUser, error) {
	user, err := spotify.GetMe(token)
	if err != nil {
		return nil, err
	}

	db, ok := dbInstance.(*sql.DB)
	if !ok {
		return nil, model.ErrFailedGetDB
	}

	err = database.SaveAccessToken(db, token, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// setAuthenticationData アプリケーションタイプに応じてJWTまたはセッションデータを設定する
func setAuthenticationData(c *gin.Context, userID string, isNativeApp bool) error {
	if isNativeApp {
		return setJWTData(c, userID)
	}
	return setSessionData(c, userID)
}

// setJWTData ネイティブアプリ用にJWTトークンを生成・設定する
func setJWTData(c *gin.Context, userID string) error {
	jwtToken, err := utils.GenerateJWT(userID)
	if err != nil {
		return err
	}

	c.Set("jwtToken", jwtToken)
	c.Set("userId", userID)
	return nil
}

// setSessionData Webアプリ用にセッションデータを設定する
func setSessionData(c *gin.Context, userID string) error {
	session := sessions.Default(c)
	session.Set("userId", userID)
	return session.Save()
}
