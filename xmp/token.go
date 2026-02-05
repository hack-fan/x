package xmp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// accessTokenServer implements the AccessTokenServer interface.
type accessTokenServer struct {
	appID  string
	secret string
	key    string
	rest   *resty.Client
	rdb    *redis.Client
	log    *zap.SugaredLogger
	ctx    context.Context
}

// newAccessTokenServer creates a new accessTokenServer, uses util.DefaultHttpClient if httpClient == nil.
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

// Required by this wechat package, weird
func (s *accessTokenServer) IID01332E16DF5011E5A9D5A4DB30FED8E1() {}

// Token gets token from cache or WeChat server
func (s *accessTokenServer) Token() (string, error) {
	token, err := s.rdb.Get(s.ctx, s.key).Result()
	if err == redis.Nil {
		// Catch and print the error again, mp package is unreliable and doesn't print
		token, err = s.requestToken()
		if err != nil {
			s.log.Errorf("mp token server error: %s", err)
		}
	} else if err != nil {
		s.log.Errorf("mp token server redis error:%s", err)
		return "", err
	}
	return token, nil
}

// This interface is useless, don't know what the author was thinking, force refresh the token
func (s *accessTokenServer) RefreshToken(current string) (token string, err error) {
	s.log.Infow("refresh mp token", "current", current)
	return s.requestToken()
}

type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
	ErrorCode int    `json:"errcode"`
	ErrorMsg  string `json:"errmsg"`
}

// updateToken gets new access_token from WeChat server, stores it in cache, and returns the access_token.
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
	body := resp.Body()
	s.log.Infof("mp token body: %s", string(body))
	at := new(accessToken)
	err = json.Unmarshal(body, at)
	if err != nil {
		return "", err
	}
	// check mp error
	if at.ErrorCode != 0 {
		return "", errors.New(at.ErrorMsg)
	}
	// last check
	if at.Token == "" || at.ExpiresIn == 0 {
		return "", errors.New("get mp token unknown error")
	}
	// set to redis
	err = s.rdb.Set(s.ctx, s.key, at.Token, time.Second*time.Duration(at.ExpiresIn)).Err()
	if err != nil {
		return "", err
	}

	return at.Token, nil
}
