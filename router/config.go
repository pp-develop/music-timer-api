package router

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// setupCORS configures CORS settings based on environment
func setupCORS(router *gin.Engine) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		log.Fatalln("ENVIRONMENT variable must be set (production or development)")
	}
	isProduction := environment == "production"

	allowedOrigins := []string{}

	// Web用URL（環境に関わらず必須）
	baseURL := os.Getenv("BASE_URL")
	apiURL := os.Getenv("API_URL")

	if baseURL != "" {
		allowedOrigins = append(allowedOrigins, baseURL)
	}
	if apiURL != "" {
		allowedOrigins = append(allowedOrigins, apiURL)
	}

	// 開発環境のみネイティブアプリのURLを追加（テスト用）
	if !isProduction {
		nativeBaseURL := os.Getenv("NATIVE_BASE_URL")
		nativeAPIURL := os.Getenv("NATIVE_API_URL")

		if nativeBaseURL != "" {
			allowedOrigins = append(allowedOrigins, nativeBaseURL)
		}
		if nativeAPIURL != "" {
			allowedOrigins = append(allowedOrigins, nativeAPIURL)
		}
	}

	// 最低1つのOriginが設定されていることを確認
	if len(allowedOrigins) == 0 {
		log.Fatalln("No valid origins configured for CORS. Check BASE_URL and API_URL environment variables.")
	}

	log.Printf("[INFO] CORS configuration - Environment: %s, AllowedOrigins: %v",
		environment, allowedOrigins)

	router.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			"POST",
			"GET",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Access-Control-Allow-Credentials",
			"Access-Control-Allow-Headers",
			"Access-Control-Allow-Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}))
}

// setupSession configures session middleware based on environment
func setupSession(router *gin.Engine) {
	environment := os.Getenv("ENVIRONMENT")
	isProduction := environment == "production"

	// Cookie configuration: Secure and SameSite based on environment
	// Development: Secure=false, SameSite=Lax (allows HTTP)
	// Production: Secure=true, SameSite=None (HTTPS only, cross-site allowed)
	var sameSiteMode http.SameSite
	if isProduction {
		sameSiteMode = http.SameSiteNoneMode // Cross-site allowed (requires HTTPS)
	} else {
		sameSiteMode = http.SameSiteLaxMode // Normal mode (works with HTTP)
	}

	store := cookie.NewStore([]byte(os.Getenv("COOKIE_SECRET")))
	store.Options(sessions.Options{
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: sameSiteMode,
		Path:     "/",
	})

	log.Printf("[INFO] Session configuration - Environment: %s, Secure: %v, SameSite: %v",
		os.Getenv("ENVIRONMENT"), isProduction, sameSiteMode)
	router.Use(sessions.Sessions("mysession", store))
}
