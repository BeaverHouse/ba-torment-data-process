package common

import (
	"go.uber.org/zap"
)

var logger *zap.Logger

func InitLogger() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func LogInfo(message string, fields ...zap.Field) {
	logger.Info(message, fields...)
}

func LogError(err error) {
	if err == nil {
		return
	}

	if runtimeErr, ok := err.(*RuntimeError); ok {
		logger.Error(runtimeErr.Message, zap.String("function", runtimeErr.FunctionName))
		return
	}

	logger.Error(err.Error())
}
