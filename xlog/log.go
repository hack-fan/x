package xlog

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Hook can process log before output
type Hook interface {
	Process(entry zapcore.Entry) error
}

var httpc = &http.Client{Timeout: time.Second * 30}

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
