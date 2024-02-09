package l2

import (
	"context"
	"github.com/thnthien/great-deku/l2/sentry"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	RequestIDCtxKey = "requestid"
)

type LoggerOption func(logger *Logger)

func WithRequestIDCtxKey(requestIDCtxKey string) LoggerOption {
	return func(logger *Logger) {
		logger.requestIDCtxKey = requestIDCtxKey
	}
}

func WithLogLevel(level string) LoggerOption {
	return func(logger *Logger) {
		err := logger.logLevel.UnmarshalText([]byte(level))
		if err != nil {
			log.Panicf("invalid log level: %s", err)
		}
	}
}

func WithZapOptions(opts []zap.Option) LoggerOption {
	return func(logger *Logger) {
		logger.zapOpts = opts
	}
}

func NewWithSentry(sentryCfg *sentry.Configuration) LoggerOption {
	return func(logger *Logger) {
		logger.sentryCfg = sentryCfg
	}
}

// Logger wraps zap.Logger
type Logger struct {
	*zap.Logger

	zapOpts         []zap.Option
	sentryCfg       *sentry.Configuration
	logLevel        Level
	requestIDCtxKey string
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	logger := l.Logger.With(fields...)
	return &Logger{
		Logger:          logger,
		zapOpts:         l.zapOpts,
		sentryCfg:       l.sentryCfg,
		logLevel:        l.logLevel,
		requestIDCtxKey: l.requestIDCtxKey,
	}
}

func (l *Logger) Trace(msg string, fields ...zap.Field) {
	l.Log(TraceLevel, msg, fields...)
}

func (l *Logger) TraceCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.LogCtx(ctx, TraceLevel, msg, fields...)
}

func (l *Logger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.LogCtx(ctx, zapcore.DebugLevel, msg, fields...)
}

func (l *Logger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.LogCtx(ctx, zapcore.InfoLevel, msg, fields...)
}

func (l *Logger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.LogCtx(ctx, zapcore.WarnLevel, msg, fields...)
}

func (l *Logger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.LogCtx(ctx, zapcore.ErrorLevel, msg, fields...)
}

func (l *Logger) LogCtx(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	rid := l.getRequestID(ctx)
	if rid != "" {
		fields = append(fields, String("request_id", rid))
	}

	var log func(string, ...zap.Field)

	switch level {
	case TraceLevel:
		log = l.Trace
	case zapcore.DebugLevel:
		log = l.Debug
	case zapcore.InfoLevel:
		log = l.Info
	case zapcore.WarnLevel:
		log = l.Warn
	case zapcore.ErrorLevel:
		log = l.Error
	case zapcore.DPanicLevel:
		log = l.DPanic
	case zapcore.PanicLevel:
		log = l.Panic
	case zapcore.FatalLevel:
		log = l.Fatal
	case zapcore.InvalidLevel:
		log = l.Panic
	}

	log(msg, fields...)
}

func (l *Logger) getRequestID(ctx context.Context) string {
	rid, _ := ctx.Value(l.requestIDCtxKey).(string)
	return rid
}
