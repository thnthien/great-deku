package l2

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/k0kubun/pp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/thnthien/great-deku/l2/config"
)

const (
	ConsoleEncoderName string        = "custom_console"
	TraceLevel         zapcore.Level = -2
)

var (
	cfgMtx      = sync.Mutex{}
	colorEnable = false
)

// SetColorEnable force set value to colorEnable
func SetColorEnable(enable bool) {
	cfgMtx.Lock()
	defer cfgMtx.Unlock()
	colorEnable = enable
}

func init() {
	cEnable := os.Getenv("LOG_COLOR")
	colorEnable = strings.ToLower(cEnable) == "true"

	err := zap.RegisterEncoder(ConsoleEncoderName, func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		return NewConsoleEncoder(cfg), nil
	})
	if err != nil {
		panic(err)
	}
}

type Level struct {
	zapcore.Level
}

func (l *Level) UnmarshalText(text []byte) error {
	s := string(bytes.ToLower(text))
	if s == "trace" {
		l.Level = TraceLevel
		return nil
	}
	return l.Level.UnmarshalText(text)
}

type dd struct {
	v interface{}
}

func (d dd) String() string {
	return pp.Sprint(d.v)
}

// Dump renders object for debugging
func Dump(v interface{}) fmt.Stringer {
	return dd{v}
}

// region logfields
// Shorthand functions for logging.
var (
	Any        = zap.Any
	Bool       = zap.Bool
	Duration   = zap.Duration
	Float64    = zap.Float64
	Int        = zap.Int
	Int64      = zap.Int64
	Skip       = zap.Skip
	String     = zap.String
	Stringer   = zap.Stringer
	Time       = zap.Time
	Uint       = zap.Uint
	Uint32     = zap.Uint32
	Uint64     = zap.Uint64
	Uintptr    = zap.Uintptr
	ByteString = zap.ByteString
	Error      = zap.Error
)

// Stack print stack trace
func Stack() zapcore.Field {
	return zap.Stack("stack")
}

// Int32 print int32 value
func Int32(key string, val int32) zapcore.Field {
	return zap.Int(key, int(val))
}

// Object print object as colored if color is enable
func Object(key string, val interface{}) zapcore.Field {
	if colorEnable {
		return zap.Stringer(key, Dump(val))
	}
	return zap.Any(key, val)
}

//endregion

func NewLogger(opts ...LoggerOption) Logger {
	ll := Logger{}
	err := ll.logLevel.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	if err != nil {
		log.Panicf("error when init log level: %s", err)
	}

	for _, opt := range opts {
		opt(&ll)
	}

	loggerConfig := config.Configuration{
		Config: zap.Config{
			Level:            zap.NewAtomicLevelAt(ll.logLevel.Level),
			Development:      false,
			Encoding:         ConsoleEncoderName,
			EncoderConfig:    DefaultConsoleEncoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
		Sentry: ll.sentryCfg,
	}
	stacktraceLevel := zap.NewAtomicLevelAt(zapcore.PanicLevel)

	buildOpts := append(ll.zapOpts, zap.AddStacktrace(stacktraceLevel))
	zLogger, err := loggerConfig.Build(buildOpts...)
	if err != nil {
		log.Panicf("error when build zap logger from build options: %s", err)
	}
	ll.Logger = zLogger

	return ll
}
