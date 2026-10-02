package rdb

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func SetLogger(logger *zap.SugaredLogger) {
	log = logger
}

type Config struct {
	Host     string `default:"redis"`
	Port     string `default:"6379"`
	Password string
	DB       int `default:"0"`
}

func New(config Config) *redis.Client {
	var kv = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if log == nil {
		logger, _ := zap.NewDevelopment()
		log = logger.Sugar()
	}

	var err error
	for i := range 60 {
		if i > 0 {
			time.Sleep(time.Second)
		}
		if err = kv.Ping(context.Background()).Err(); err == nil {
			log.Info("redis connect successful")
			return kv
		}
	}
	panic(fmt.Sprintf("connect to redis failed: %v", err))
}
