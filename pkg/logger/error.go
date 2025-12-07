package logger

import (
	"log/slog"
	"os"
)

// defaultLogger はアプリケーション全体で使用するデフォルトのロガー
var defaultLogger *slog.Logger

func init() {
	// JSON形式のハンドラーを使用（本番環境向け）
	// 開発環境では TextHandler に変更可能
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// Info は情報レベルのログを出力する
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Warn は警告レベルのログを出力する
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error はエラーレベルのログを出力する
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// Debug はデバッグレベルのログを出力する
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}
