package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getJWTSecret())

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}
	// 256ビット = 32バイト以上を推奨
	if len(secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters long for security")
	}
	return secret
}

// GenerateJWT は指定されたユーザーIDに対して新しいJWTトークンを生成します。
//
// このトークンは24時間有効で、以下の情報を含みます：
// - ユーザーID（クレーム内）
// - 発行時刻（IssuedAt）
// - 有効期限（ExpiresAt）
//
// Parameters:
//   - userID: JWTトークンに埋め込むユーザーID（必須）
//
// Returns:
//   - string: 生成されたJWTトークン文字列（成功時）
//   - error: トークン生成に失敗した場合のエラー（JWT_SECRET環境変数が未設定の場合など）
//
// Example:
//
//	token, err := GenerateJWT("user123")
//	if err != nil {
//	    log.Printf("JWT生成エラー: %v", err)
//	    return
//	}
//	// tokenを使用してAPIレスポンスに含める
func GenerateJWT(userID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24時間有効
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the user ID
func ValidateJWT(tokenString string) (string, error) {
	claims := &Claims{}
	
	// パーサーオプションでクロックスキューを設定（5分の猶予）
	parser := jwt.NewParser(
		jwt.WithLeeway(5 * time.Minute),
		jwt.WithValidMethods([]string{"HS256"}), // 許可する署名メソッドを明示的に指定
	)
	
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 署名メソッドの検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// 具体的にHS256であることを確認
		if token.Method.Alg() != "HS256" {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	// ユーザーIDが空でないことを確認
	if claims.UserID == "" {
		return "", errors.New("user ID is empty in token")
	}

	return claims.UserID, nil
}
