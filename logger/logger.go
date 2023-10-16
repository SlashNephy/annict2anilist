package logger

import (
	"log/slog"
	"os"
)

var level slog.LevelVar

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &level,
	}))
	slog.SetDefault(logger)
}

func SetLevel(value string) {
	_ = level.UnmarshalText([]byte(value))
}
