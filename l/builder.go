package l

import (
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/thnthien/great-deku/l/config"
	"github.com/thnthien/great-deku/l/sentry"
	"github.com/thnthien/great-deku/l/telegram"
)

type Builder struct {
	Sentry   *sentry.Configuration `yaml:"sentry"`
	File     *lumberjack.Logger    `yaml:"file"`
	Telegram *telegram.Config      `yaml:"telegram"`
}

func (b Builder) Build(opts ...zap.Option) Logger {
	var name string
	if name == "" {
		_, filename, _, _ := runtime.Caller(1)
		name = filepath.Dir(truncFilename(filename))
	}

	var enabler zap.AtomicLevel
	if e, ok := enablers[name]; ok {
		enabler = e
	} else {
		enabler = zap.NewAtomicLevel()
		enablers[name] = enabler
	}

	setLogLevelFromEnv(name, enabler)

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
	return Logger{logger, logger.Sugar()}
}
