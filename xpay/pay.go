package xpay

import (
	"fmt"
	"net/http"

	"github.com/hyacinthus/wechat/mch/core"
	"github.com/labstack/echo/v4"
)

// Config is wechat pay config
type Config struct {
	MPAppID     string
	MchID       string
	APIKey      string
	CertPath    string `default:"/run/cert/apiclient_cert.pem"`
	CertKeyPath string `default:"/run/cert/apiclient_key.pem"`
}

// NewClient create a wechat pay client
func NewClient(config Config) *core.Client {
	httpc, err := core.NewTLSHttpClient(config.CertPath, config.CertKeyPath)
	if err != nil {
		panic(fmt.Errorf("pay tls client failed:%w", err))
	}
	return core.NewClient(config.MPAppID, config.MchID, config.APIKey, httpc)
}

// Handler will handle wechat pay callbacks
type Handler interface {
	ServeMsg(ctx *core.Context)
	ServeError(w http.ResponseWriter, r *http.Request, err error)
}

// NewEchoHandler gen a echo handler
func NewEchoHandler(config Config, h Handler) echo.HandlerFunc {
	s := core.NewServer(config.MPAppID, config.MchID, config.APIKey, h, h)
	return func(c echo.Context) error {
		s.ServeHTTP(c.Response().Writer, c.Request(), c.QueryParams())
		return nil
	}
}
