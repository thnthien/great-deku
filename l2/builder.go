package l2

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/thnthien/great-deku/l2/config"
	"github.com/thnthien/great-deku/l2/sentry"
	"github.com/thnthien/great-deku/l2/telegram"
)

type Builder struct {
	Sentry   *sentry.Configuration `yaml:"sentry"`
	File     *lumberjack.Logger    `yaml:"file"`
	Telegram *telegram.Config      `yaml:"telegram"`
}

func (b Builder) Build(opts ...zap.Option) Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	var lv Level
	err := lv.UnmarshalText([]byte(logLevel))
	if err != nil {
		lv.Level = zapcore.InfoLevel
	}

	enabler := zap.NewAtomicLevelAt(lv.Level)

	loggerConfig := config.Configuration{
		Config: zap.Config{
			Level:            enabler,
			Development:      false,
			Encoding:         ConsoleEncoderName,
			EncoderConfig:    DefaultConsoleEncoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
		Sentry:   b.Sentry,
		File:     b.File,
		Telegram: b.Telegram,
	}

	stacktraceLevel := zap.NewAtomicLevelAt(zapcore.PanicLevel)

	opts = append(opts, zap.AddStacktrace(stacktraceLevel))
	logger, err := loggerConfig.Build(opts...)
	if err != nil {
		panic(err)
	}
	return Logger{Logger: logger}
}
