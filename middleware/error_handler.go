package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pp-develop/music-timer-api/model"
	"github.com/pp-develop/music-timer-api/pkg/logger"
)

// ErrorHandlerMiddleware は全てのエンドポイントで共通のエラー処理を行うミドルウェア
// ハンドラーがエラーをコンテキストに設定した場合、適切なHTTPレスポンスに変換します
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // ハンドラーを実行

		// レスポンスが既に書き込まれている場合はスキップ
		if c.Writer.Written() {
			return
		}

		// コンテキストからエラーを取得
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

func handleError(c *gin.Context, err error) {
	log.Printf("[ERROR-HANDLER] Handling error: %v (Type: %T)", err, err)
	logger.LogError(err)

	// 認証エラー
	if errors.Is(err, model.ErrAccessTokenExpired) {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code: model.CodeTokenExpired,
		})
		return
	}

	// リソース不足エラー
	if errors.Is(err, model.ErrNotFoundTracks) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTracksNotFound,
		})
		return
	}

	if errors.Is(err, model.ErrNotEnoughTracks) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutInsufficientTracks,
		})
		return
	}

	if errors.Is(err, model.ErrTimeoutCreatePlaylist) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeTimeoutNoMatch,
		})
		return
	}

	if errors.Is(err, model.ErrNoFavoriteTracks) {
		c.JSON(http.StatusNotFound, model.ErrorResponse{
			Code: model.CodeNoFavoriteTracks,
		})
		return
	}

	// Spotify API制限エラー
	if errors.Is(err, model.ErrSpotifyRateLimit) {
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
			Code: model.CodeSpotifyRateLimit,
		})
		return
	}

	if errors.Is(err, model.ErrPlaylistQuotaExceeded) {
		c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
			Code: model.CodePlaylistQuotaExceeded,
		})
		return
	}

	// 処理エラー
	if errors.Is(err, model.ErrPlaylistCreationFailed) {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Code: model.CodePlaylistCreationFailed,
		})
		return
	}

	if errors.Is(err, model.ErrTrackAdditionFailed) {
		c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Code: model.CodePlaylistCreationFailed,
		})
		return
	}

	// セッション/認証エラー
	// JWT/セッション認証を問わず、401 Unauthorized で統一
	if errors.Is(err, model.ErrFailedGetSession) {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Code: model.CodeTokenExpired,
		})
		return
	}

	if errors.Is(err, model.ErrNotFoundPlaylist) {
		c.Status(http.StatusNoContent)
		return
	}

	// デフォルト: 内部サーバーエラー
	c.JSON(http.StatusInternalServerError, model.ErrorResponse{
		Code: model.CodeInternalError,
	})
}
