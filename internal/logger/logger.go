package logger

import "go.uber.org/zap"

func New() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"

	return cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
}
