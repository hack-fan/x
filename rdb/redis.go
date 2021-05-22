package rdb

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
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
		Addr:     config.Host + ":" + config.Port,
		Password: config.Password,
		DB:       config.DB,
	})

	if log == nil {
		logger, _ := zap.NewDevelopment()
		log = logger.Sugar()
	}

	var i int
	for {
		err := kv.Ping(context.Background()).Err()
		if err != nil {
			if i >= 60 {
				panic("connect to redis failed")
			}
			time.Sleep(time.Second)
			continue
		}
		break
	}

	log.Info("redis connect successful")

	return kv
}
