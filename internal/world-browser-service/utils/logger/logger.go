package logger

import (
	"go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

func GetLogger() *zap.SugaredLogger {
	if Sugar == nil {
		logger, _ := zap.NewProduction()
		Sugar = logger.Sugar()
	}
	return Sugar
}

func Sync() {
	if Sugar != nil {
		Sugar.Sync()
	}
}
