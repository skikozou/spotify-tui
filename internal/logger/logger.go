package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	logFile *os.File
	enabled bool
)

// Init はロガーを初期化する
// debug が true の場合のみログを出力する
func Init(debug bool, logFilePath string) error {
	enabled = debug
	if !debug {
		// デバッグモードでない場合はログを無効化
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return nil
	}

	// ログディレクトリを作成
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// ログファイルを開く
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// slogのデフォルトロガーを設定
	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(handler))

	Info("Logger initialized", "logFile", logFilePath)
	return nil
}

// Close はログファイルを閉じる
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Debug はDEBUGレベルのログを出力する
func Debug(msg string, args ...any) {
	if enabled {
		slog.Debug(msg, args...)
	}
}

// Info はINFOレベルのログを出力する
func Info(msg string, args ...any) {
	if enabled {
		slog.Info(msg, args...)
	}
}

// Warn はWARNレベルのログを出力する
func Warn(msg string, args ...any) {
	if enabled {
		slog.Warn(msg, args...)
	}
}

// Error はERRORレベルのログを出力する
func Error(msg string, args ...any) {
	if enabled {
		slog.Error(msg, args...)
	}
}
