package main

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = NewLogger()
	if err != nil {
		panic(err)
	}
}

func NewLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.Sampling = nil
	config.Encoding = "console"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	return config.Build()
}
