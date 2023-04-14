package l

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	RequestIDCtxKey = "requestid"
)

func getRequestID(ctx context.Context) string {
	rid, _ := ctx.Value(RequestIDCtxKey).(string)
	return rid
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
	rid := getRequestID(ctx)
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
	}

	log(msg, fields...)
}
