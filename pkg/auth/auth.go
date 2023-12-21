package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify"
	"github.com/pp-develop/make-playlist-by-specify-time-api/database"
	"github.com/pp-develop/make-playlist-by-specify-time-api/model"
)

// Auth はユーザー認証を行います。
// この関数は、まずセッションからユーザーIDを取得します。
//　ユーザーIDが存在しない場合、ErrFailedGetSession エラーを返します。
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

	user, err := database.GetUser(userId)
	if err != nil {
		return err
	}

	token, err := spotify.RefreshToken(user)
	if err != nil {
		return err
	}

	err = database.UpdateAccessToken(token, user.Id)
	if err != nil {
		return err
	}

	session.Set("userId", user.Id)
	session.Save()

	return nil
}
