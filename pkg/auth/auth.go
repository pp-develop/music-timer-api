package auth

import (
	"log"
	"time"

	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/search"
	"github.com/pp-develop/music-timer-api/utils"
	"golang.org/x/oauth2"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Auth はユーザー認証を行います。
// この関数は、まずセッションからユーザーIDを取得します。
// 　ユーザーIDが存在しない場合、ErrFailedGetSession エラーを返します。
// ユーザーIDが存在する場合、データベースからユーザー情報を取得し、Spotifyのトークンを更新します。
// 最終的に、更新されたユーザーIDをセッションに保存し、セッションを保存します。
//
// この関数は、以下のエラーを返す可能性があります：
//   - model.ErrFailedGetSession: セッションからユーザーIDを取得できなかった場合。
//   - database関連のエラー: ユーザー情報の取得に失敗した場合。
//   - spotify関連のエラー: トークンの更新に失敗した場合。
//   - database関連のエラー: アクセストークンのデータベースへの更新に失敗した場合。
//
// 成功した場合はnilを返し、エラーが発生した場合はそのエラーを返します。
func Auth(c *gin.Context) error {
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return model.ErrFailedGetSession
	}
	userId := v.(string)

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return model.ErrFailedGetDB
	}

	user, err := database.GetUser(dbInstance, userId)
	if err != nil {
		return err
	}

	// トークンの有効期限チェック
	if !checkTokenExpiration(user) {
		// トークンリフレッシュ
		token, err := spotify.RefreshToken(user)
		if err != nil {
			return err
		}

		// アクセストークンの更新
		err = database.UpdateAccessToken(dbInstance, token, user.Id)
		if err != nil {
			return err
		}
	}

	// 更新日時のチェック
	// TODO:: 別リクエストで行う
	updateAt, err := time.Parse(time.RFC3339, user.UpdateAt)
	if err != nil {
		log.Printf("failed to parse UpdatedAt: %v", err)
		return err
	}

	if time.Since(updateAt).Hours() > 24 {
		err = search.SaveFavoriteTracks(dbInstance, &oauth2.Token{
			AccessToken:  user.AccessToken,
			RefreshToken: user.RefreshToken,
		}, userId)
		if err != nil {
			return err
		}

		// 更新日時を現在の時間に更新
		err = database.UpdateUserUpdateAt(dbInstance, userId, time.Now())
		if err != nil {
			return err
		}
	}

	// セッションの保存
	session.Set("userId", user.Id)
	err = session.Save()
	if err != nil {
		return err
	}
	return nil
}

func checkTokenExpiration(user model.User) bool {
	currentTime := time.Now().Unix()
	tokenExpiration := int64(user.TokenExpiration)
	return tokenExpiration > currentTime
}
