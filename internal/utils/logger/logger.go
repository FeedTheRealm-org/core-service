package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func InitLogger(production bool) *zap.SugaredLogger {
	var logger *zap.Logger
	if production {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}

	Logger = logger.Sugar() // This might be expensive for performance-critical endpoints
	return Logger
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
