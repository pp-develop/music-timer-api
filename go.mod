module github.com/pp-develop/music-timer-api

go 1.23.0

toolchain go1.24.1

require (
	github.com/gin-contrib/cors v1.7.5
	github.com/gin-contrib/sessions v1.0.3
	github.com/gin-gonic/gin v1.10.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.27.6
	github.com/zmb3/spotify/v2 v2.4.3
	golang.org/x/oauth2 v0.29.0
)

require (
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.0.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/arch v0.16.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pp-develop/music-timer-api/pkg/artist => ./pkg/artist

replace github.com/pp-develop/music-timer-api/pkg/auth => ./pkg/auth

replace github.com/pp-develop/music-timer-api/pkg/playlist => ./pkg/playlist

replace github.com/pp-develop/music-timer-api/pkg/logger => ./pkg/logger

replace github.com/pp-develop/music-timer-api/pkg/json => ./pkg/json

replace github.com/pp-develop/music-timer-api/pkg/track => ./pkg/track

replace github.com/pp-develop/music-timer-api/api/spotify => ./api/spotify

replace github.com/pp-develop/music-timer-api/router => ./router

replace github.com/pp-develop/music-timer-api/model => ./model

replace github.com/pp-develop/music-timer-api/database => ./database
