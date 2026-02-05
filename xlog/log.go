package xlog

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Hook can process log before output
type Hook interface {
	Process(entry zapcore.Entry) error
}

// Hooks is a slice of Hook that implements Hook interface
type Hooks []Hook

// Process calls each hook in sequence
func (h Hooks) Process(entry zapcore.Entry) error {
	for _, hook := range h {
		if err := hook.Process(entry); err != nil {
			return err
		}
	}
	return nil
}

var httpc = &http.Client{Timeout: time.Second * 30}

// httpPostWithContext sends HTTP POST with context and timeout
func httpPostWithContext(ctx context.Context, url string, body any) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return httpc.Do(req)
}

func New(debug bool, hooks ...Hook) *zap.Logger {
	var log *zap.Logger
	var err error
	if debug {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	for _, hook := range hooks {
		log = log.WithOptions(zap.Hooks(hook.Process))
	}
	return log
}

// Config is the logger configuration
type Config struct {
	Debug    bool
	Hooks    []Hook
	Level    zapcore.Level
	Encoding string // "json" or "console"
}

// Option configures the logger
type Option func(*Config)

// WithDebug sets debug mode
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithHooks adds hooks
func WithHooks(hooks ...Hook) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

// WithLevel sets the log level
func WithLevel(level zapcore.Level) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithEncoding sets the encoding format
func WithEncoding(encoding string) Option {
	return func(c *Config) {
		c.Encoding = encoding
	}
}

// NewWithConfig creates a logger with configuration options
func NewWithConfig(opts ...Option) *zap.Logger {
	cfg := &Config{
		Debug:    false,
		Level:    zapcore.InfoLevel,
		Encoding: "json",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var zapConfig zap.Config
	if cfg.Debug {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = cfg.Encoding
	}
	zapConfig.Level = zap.NewAtomicLevelAt(cfg.Level)

	log, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	for _, hook := range cfg.Hooks {
		log = log.WithOptions(zap.Hooks(hook.Process))
	}

	return log
}
