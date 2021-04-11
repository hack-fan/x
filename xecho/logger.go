package xecho

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Skipper func(ctx echo.Context) bool

// SkipRule must be fully equal
type SkipRule struct {
	Method     string
	Path       string
	StatusCode int
}

// NewSkipper gen a logger skipper
func NewSkipper(rules []SkipRule) Skipper {
	return func(c echo.Context) bool {
		for _, rule := range rules {
			if c.Request().Method == rule.Method &&
				c.Path() == rule.Path &&
				c.Response().Status == rule.StatusCode {
				return true
			}
		}
		return false
	}
}

// ZapLoggerWithSkipper use zap as request logger
// thank https://github.com/brpaz/echozap
func ZapLoggerWithSkipper(log *zap.Logger, skipper Skipper) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// skip log
			if skipper != nil && skipper(c) {
				return nil
			}

			req := c.Request()
			res := c.Response()

			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("time", time.Since(start).String()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
				zap.String("user_agent", req.UserAgent()),
			}

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
				fields = append(fields, zap.String("request_id", id))
			}

			// log all at info level
			// if there is a server error, echo error handler will log it as error there.
			n := res.Status
			switch {
			case n >= 500:
				log.Info("Server Error", fields...)
			case n >= 400:
				log.Info("Client Error", fields...)
			case n >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}

// ZapLogger use zap for echo logger
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return ZapLoggerWithSkipper(log, nil)
}
