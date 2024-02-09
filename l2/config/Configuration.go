package config

import (
	"errors"
	"io"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/thnthien/great-deku/l2/sentry"
	"github.com/thnthien/great-deku/l2/telegram"
)

// Configuration defines the desired logging options.
type Configuration struct {
	zap.Config

	Sentry   *sentry.Configuration `yaml:"sentry"`
	File     *lumberjack.Logger    `yaml:"file"`
	Telegram *telegram.Config      `yaml:"telegram"`
}

// Configure initializes logging configuration struct from config provider
func (c *Configuration) Configure(cfg Value) error {

	// Because log.Configuration embeds zap, the PopulateStruct
	// does not work properly as it's unable to serialize fields directly
	// into the embedded struct, so inner struct has to be treated as a
	// separate object
	//
	// first, use the default zap configuration
	zapCfg := DefaultConfiguration().Config

	// override the embedded zap.Config stuct from config
	if err := cfg.PopulateStruct(&zapCfg); err != nil {
		return errors.New("unable to parse logging config")
	}

	// use the overriden zap config
	c.Config = zapCfg

	// override any remaining things fom config, i.e. Sentry
	if err := cfg.PopulateStruct(&c); err != nil {
		return errors.New("unable to parse logging config")
	}

	return nil
}

// Build constructs a *zap.Logger with the configured parameters.
func (c Configuration) Build(opts ...zap.Option) (*zap.Logger, error) {
	logger, err := c.Config.Build(opts...)
	if err != nil {
		// If there's an error or there's no Sentry config, we don't need to do
		// anything but delegate.
		return logger, err
	}
	var cores []zapcore.Core
	if c.Sentry != nil {
		sentryObj, err := c.Sentry.Build()
		if err != nil {
			log.Printf("error when init sentry log: %s", err)
		}
		cores = append(cores, sentryObj)
	}
	if c.File != nil {
		w := zapcore.AddSync(c.File)
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(c.EncoderConfig), w, c.Config.Level)
		cores = append(cores, core)
	}
	if c.Telegram != nil {
		var writer io.Writer
		if c.Telegram.Bot != nil {
			writer = telegram.NewWithBot(c.Telegram.Bot, c.Telegram.ChatID)
		} else {
			writer, err = telegram.New(c.Telegram.Token, c.Telegram.ChatID)
			if err != nil {
				log.Printf("error when init telegram log: %s", err)
			}
		}
		w := zapcore.AddSync(writer)
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(c.EncoderConfig), w, zap.NewAtomicLevelAt(zapcore.ErrorLevel))
		cores = append(cores, core)
	}
	if len(cores) == 0 {
		return logger, err
	}
	return logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		cores = append([]zapcore.Core{core}, cores...)
		return zapcore.NewTee(cores...)
	})), nil
}

// DefaultConfiguration returns a fallback configuration for applications that
// don't explicitly configure logging.
func DefaultConfiguration() Configuration {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}

	return Configuration{
		Config: cfg,
	}
}
