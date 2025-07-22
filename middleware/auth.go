package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/utils"
)

// JWTAuthMiddleware は JWT トークンによる認証を強制するミドルウェアです
// 認証が必須のAPIエンドポイントで使用します
// Authorization ヘッダーから Bearer トークンを抽出し、JWT の有効性を検証します
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// ヘッダーが存在しない場合は 401 エラーを返す
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Bearer トークンから JWT を抽出
		// 例: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." → "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// "Bearer " プレフィックスがない場合は無効なフォーマット
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// JWT トークンを検証してユーザーIDを取得
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			// トークンが無効または期限切れの場合
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 認証成功: コンテキストにユーザーIDを設定して次のハンドラーに進む
		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware は JWT とセッション認証の両方に対応するミドルウェアです
// 認証情報がない場合でもリクエストを中断せず、あれば userID を設定します
// 認証されていなくても動作するが、認証情報があれば活用するエンドポイントで使用します
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// まず JWT 認証を試行
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Authorization ヘッダーが存在する場合
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				// 正しい Bearer 形式の場合、JWT を検証
				userID, err := utils.ValidateJWT(tokenString)
				if err == nil {
					// JWT 認証成功: ユーザーIDと認証タイプを設定
					c.Set("userID", userID)
					c.Set("authType", "jwt")
				}
				// JWT が無効でもエラーにはせず、処理を継続
			}
		}

		// JWT がない場合、userID はセッションミドルウェアによって設定される可能性があります
		// セッションミドルウェアはこのミドルウェアより前に実行される必要があります

		// 認証状況に関わらず次のハンドラーに進む
		c.Next()
	}
}
