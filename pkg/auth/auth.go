package auth

import (
	"time"

	"github.com/pp-develop/music-timer-api/api/spotify"
	"github.com/pp-develop/music-timer-api/database"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/utils"

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
func Auth(c *gin.Context) (model.User, error) {
	var user model.User
	session := sessions.Default(c)
	v := session.Get("userId")
	if v == nil {
		return user, model.ErrFailedGetSession
	}
	userId := v.(string)

	dbInstance, ok := utils.GetDB(c)
	if !ok {
		return user, model.ErrFailedGetDB
	}

	user, err := database.GetUser(dbInstance, userId)
	if err != nil {
		return user, err
	}

	// トークンの有効期限チェック
	if !checkTokenExpiration(user) {
		// トークンリフレッシュ
		token, err := spotify.RefreshToken(user)
		if err != nil {
			return user, err
		}

		// アクセストークンの更新
		err = database.UpdateAccessToken(dbInstance, token, user.Id)
		if err != nil {
			return user, err
		}
	}

	// セッションの保存
	session.Set("userId", user.Id)
	err = session.Save()
	if err != nil {
		return user, err
	}
	return user, nil
}

func checkTokenExpiration(user model.User) bool {
	currentTime := time.Now().Unix()
	tokenExpiration := int64(user.TokenExpiration)
	return tokenExpiration > currentTime
}
