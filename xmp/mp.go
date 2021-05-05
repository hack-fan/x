package xmp

import (
	"net/http"
	"time"

	"github.com/chanxuehong/wechat/mp/core"
	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Config mp config
type Config struct {
	AppID  string
	Secret string
	OriID  string
	Token  string
	AesKey string
}

// NewClient mp client
func NewClient(rdb *redis.Client, rest *resty.Client, log *zap.SugaredLogger, config Config) *core.Client {
	return core.NewClient(newAccessTokenServer(config.AppID, config.Secret, rdb, rest, log),
		&http.Client{Timeout: 30 * time.Second})
}

// NewServer mp server
func NewServer(mux core.Handler, config Config) *core.Server {
	return core.NewServer(config.OriID, config.AppID, config.Token, config.AesKey, mux, nil)
}
