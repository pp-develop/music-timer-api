package logger

import (
	"log/slog"
	"os"
)

func init() {
	// JSON形式のハンドラーを使用（本番環境向け）
	// 開発環境では TextHandler に変更可能
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}
