package utils

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"gitlab.crja72.ru/gospec/go8/payment/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GooseZapLogger логгер для гуся
func GooseZapLogger(logger *zap.Logger) goose.Logger {
	return &gooseZapLogger{logger: logger}
}

type gooseZapLogger struct {
	logger *zap.Logger
}

func (z gooseZapLogger) Fatalf(format string, v ...interface{}) {
	z.logger.Fatal(fmt.Sprintf(format, v...))
}

func (z gooseZapLogger) Printf(format string, v ...interface{}) {
	z.logger.Debug(fmt.Sprintf(format, v...))
}

func (z gooseZapLogger) Infof(format string, v ...interface{}) {
	z.logger.Info(fmt.Sprintf(format, v...))
}

func (z gooseZapLogger) Warnf(format string, v ...interface{}) {
	z.logger.Warn(fmt.Sprintf(format, v...))
}

func (z gooseZapLogger) Errorf(format string, v ...interface{}) {
	z.logger.Error(fmt.Sprintf(format, v...))
}

// NewLogger красивый логгер
func NewLogger(cfg *config.Config) *zap.Logger {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalColorLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logger
}
