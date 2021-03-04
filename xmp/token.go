package xmp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// accessTokenServer 实现了 AccessTokenServer 接口.
type accessTokenServer struct {
	appID  string
	secret string
	key    string
	rest   *resty.Client
	rdb    *redis.Client
	log    *zap.SugaredLogger
	ctx    context.Context
}

// newAccessTokenServer 创建一个新的 accessTokenServer, 如果 httpClient == nil 则默认使用 util.DefaultHttpClient.
func newAccessTokenServer(appID, secret string, rdb *redis.Client, rest *resty.Client, log *zap.SugaredLogger) *accessTokenServer {
	return &accessTokenServer{
		appID:  url.QueryEscape(appID),
		secret: url.QueryEscape(secret),
		key:    "mp:token:" + appID,
		rest:   rest,
		rdb:    rdb,
		log:    log,
		ctx:    context.Background(),
	}
}

// 这个 wechat 包需要，奇葩
func (s *accessTokenServer) IID01332E16DF5011E5A9D5A4DB30FED8E1() {}

// Token 从缓存或微信服务器获得token
func (s *accessTokenServer) Token() (string, error) {
	token, err := s.rdb.Get(s.ctx, s.key).Result()
	if err == redis.Nil {
		return s.requestToken()

	} else if err != nil {
		return "", err
	}
	return token, nil
}

// 这个接口就没用，不知道这作者怎么想的，强制刷新一下token吧
func (s *accessTokenServer) RefreshToken(current string) (token string, err error) {
	s.log.Infow("refresh mp token", "current", current)
	return s.requestToken()
}

type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
}

// updateToken 从微信服务器获取新的 access_token 并存入缓存, 同时返回该 access_token.
func (s *accessTokenServer) requestToken() (string, error) {
	target := "https://api.weixin.qq.com/cgi-bin/token"
	resp, err := s.rest.R().SetQueryParams(map[string]string{
		"grant_type": "client_credential",
		"appid":      s.appID,
		"secret":     s.secret,
	}).Get(target)
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("get mp token error:%s", resp.Status())
	}
	// parse
	at := new(accessToken)
	err = json.Unmarshal(resp.Body(), at)
	if err != nil {
		return "", err
	}
	// set to redis
	err = s.rdb.Set(s.ctx, s.key, at.Token, time.Second*time.Duration(at.ExpiresIn)).Err()
	if err != nil {
		return "", err
	}

	return at.Token, nil
}
