package utils

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret = []byte(getJWTSecret())

type Claims struct {
	UserID  string `json:"user_id"`
	Service string `json:"service"` // "spotify", "soundcloud", etc.
	Type    string `json:"type"`    // "access" or "refresh"
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID  string `json:"user_id"`
	Service string `json:"service"` // "spotify", "soundcloud", etc.
	JTI     string `json:"jti"`     // JWT ID for tracking
	Type    string `json:"type"`    // "refresh"
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func getJWTSecret() string {
	err := godotenv.Load()
	if err != nil {
		slog.Debug("error loading .env file", slog.Any("error", err))
	}

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

// GenerateAccessToken generates a short-lived access token (1 hour)
func GenerateAccessToken(userID string, service string) (string, error) {
	if service == "" {
		return "", errors.New("service is required")
	}

	expirationTime := time.Now().Add(1 * time.Hour) // 1時間有効
	claims := &Claims{
		UserID:  userID,
		Service: service,
		Type:    "access",
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

// GenerateRefreshToken generates a long-lived refresh token (30 days)
func GenerateRefreshToken(userID string, service string, jti string) (string, error) {
	if service == "" {
		return "", errors.New("service is required")
	}

	expirationTime := time.Now().Add(30 * 24 * time.Hour) // 30日有効
	claims := &RefreshClaims{
		UserID:  userID,
		Service: service,
		JTI:     jti,
		Type:    "refresh",
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

// GenerateTokenPair generates both access and refresh tokens
func GenerateTokenPair(userID string, service string, jti string) (*TokenPair, error) {
	accessToken, err := GenerateAccessToken(userID, service)
	if err != nil {
		return nil, err
	}

	refreshToken, err := GenerateRefreshToken(userID, service, jti)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour in seconds
		TokenType:    "Bearer",
	}, nil
}

// ValidateJWT validates a JWT token and returns the user ID and service
func ValidateJWT(tokenString string) (string, string, error) {
	claims := &Claims{}

	// パーサーオプションでクロックスキューを設定（5分の猶予）
	parser := jwt.NewParser(
		jwt.WithLeeway(5*time.Minute),
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
		return "", "", err
	}

	if !token.Valid {
		return "", "", errors.New("invalid token")
	}

	// ユーザーIDが空でないことを確認
	if claims.UserID == "" {
		return "", "", errors.New("user ID is empty in token")
	}

	// アクセストークンであることを確認
	if claims.Type != "access" {
		return "", "", errors.New("invalid token type")
	}

	// サービスが空でないことを確認
	if claims.Service == "" {
		return "", "", errors.New("service is required in token")
	}

	return claims.UserID, claims.Service, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID, service, and JTI
func ValidateRefreshToken(tokenString string) (string, string, string, error) {
	claims := &RefreshClaims{}

	// パーサーオプションでクロックスキューを設定（5分の猶予）
	parser := jwt.NewParser(
		jwt.WithLeeway(5*time.Minute),
		jwt.WithValidMethods([]string{"HS256"}),
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
		return "", "", "", err
	}

	if !token.Valid {
		return "", "", "", errors.New("invalid token")
	}

	// ユーザーIDとJTIが空でないことを確認
	if claims.UserID == "" {
		return "", "", "", errors.New("user ID is empty in token")
	}
	if claims.JTI == "" {
		return "", "", "", errors.New("JTI is empty in token")
	}

	// リフレッシュトークンであることを確認
	if claims.Type != "refresh" {
		return "", "", "", errors.New("invalid token type")
	}

	// サービスが空でないことを確認
	if claims.Service == "" {
		return "", "", "", errors.New("service is required in token")
	}

	return claims.UserID, claims.Service, claims.JTI, nil
}
