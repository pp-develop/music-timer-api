module github.com/pp-develop/make-playlist-by-specify-time-api

go 1.19

require (
	github.com/gin-contrib/cors v1.4.0
	github.com/gin-contrib/sessions v0.0.5
	github.com/gin-gonic/gin v1.9.1
	github.com/go-sql-driver/mysql v1.7.0
	github.com/google/uuid v1.3.0
	github.com/joho/godotenv v1.4.0
	github.com/urfave/cli/v2 v2.23.7
	github.com/zmb3/spotify/v2 v2.3.0
	golang.org/x/oauth2 v0.3.0
)

require (
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/gabriel-vasile/mimetype v1.4.2 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.14.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.4 // indirect
	github.com/leodido/go-urn v1.2.4 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.0.8 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/pp-develop/make-playlist-by-specify-time-api/pkg/artist => ./pkg/artist

replace github.com/pp-develop/make-playlist-by-specify-time-api/pkg/auth => ./pkg/auth

replace github.com/pp-develop/make-playlist-by-specify-time-api/pkg/playlist => ./pkg/playlist

replace github.com/pp-develop/make-playlist-by-specify-time-api/pkg/json => ./pkg/json

replace github.com/pp-develop/make-playlist-by-specify-time-api/pkg/track => ./pkg/track

replace github.com/pp-develop/make-playlist-by-specify-time-api/api/spotify => ./api/spotify

replace github.com/pp-develop/make-playlist-by-specify-time-api/router => ./router

replace github.com/pp-develop/make-playlist-by-specify-time-api/model => ./model

replace github.com/pp-develop/make-playlist-by-specify-time-api/database => ./database
