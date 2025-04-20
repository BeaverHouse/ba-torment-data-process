package common

import (
	"os"

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

// ExitIfError exits the program if error is fired.
func ExitIfError(err error) {
	if err == nil {
		return
	}

	if runtimeErr, ok := err.(*RuntimeError); ok {
		logger.Fatal(runtimeErr.Message, zap.String("function", runtimeErr.FunctionName))
		os.Exit(1)
	}

	logger.Fatal(err.Error())
}

func LogInfo(message string, fields ...zap.Field) {
	logger.Info(message, fields...)
}

func LogWarn(message string, fields ...zap.Field) {
	logger.Warn(message, fields...)
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
