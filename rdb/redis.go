package rdb

import (
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

	// TODO: ping redis

	log.Info("Redis connect successful.")

	return kv
}
