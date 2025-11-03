package logger

import (
	"go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

func InitLogger(production bool) *zap.SugaredLogger {
	var logger *zap.Logger
	if production {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}

	Sugar = logger.Sugar() // This might be expensive for performance-critical endpoints
	return Sugar
}

func GetLogger() *zap.SugaredLogger {
	if Sugar == nil {
		InitLogger(true)
	}
	return Sugar
}

func Sync() {
	if Sugar != nil {
		_ = Sugar.Sync()
	}
}
