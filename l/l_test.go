package l

import (
	"context"
	"errors"
	"testing"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/thnthien/great-deku/l/telegram"
)

func TestNew(t *testing.T) {

	ll = Builder{
		File: &lumberjack.Logger{
			Filename:   "logs/logs.log",
			MaxSize:    1,
			MaxAge:     7,
			MaxBackups: 10,
		},
	}.Build()
	//ll = NewWithSentry(&sentry.Configuration{
	//	DSN:   "https://6c823523782944c597fcc102c8b6ae4e@o390151.ingest.sentry.io/5231166",
	//	Trace: struct{ Disabled bool }{Disabled: false},
	//})
	defer ll.Sync()
	a := map[string]interface{}{
		"testdebug": 1,
	}
	ll.Trace("test trace", Any("test trace", a))
	ll.Debug("test debug", Any("test debug", a))
	ll.Info("test info", Any("test debug", a))
	ll.Warn("test warn")
	//ll.Panic("fatal")
	ll.Error("test err")

	ctx := context.WithValue(context.Background(), RequestIDCtxKey, "test_id")
	ll.DebugCtx(ctx, "test request_id", String("message", "test"))
	ll.InfoCtx(ctx, "test request_id", String("message", "test"))
	ll.WarnCtx(ctx, "test request_id", String("message", "test"))
	ll.ErrorCtx(ctx, "test request_id", String("message", "test"))
}

func Test_TelegramLog(t *testing.T) {
	ll = Builder{Telegram: &telegram.Config{
		Token:  "",
		ChatID: 0,
	}}.Build()
	ll.Info("test info")
	ll.Error("test error", Error(errors.New("this error is for testing")))
}
