package xecho

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerSkipper skip the heartbeat /status log
func LoggerSkipper(c echo.Context) bool {
	if c.Path() == "/status" {
		return true
	}
	return false
}

// LoggerMid skip /status endpoint, will be deprecated.
func LoggerMid() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: LoggerSkipper,
	})
}

// ZapLogger use zap as request logger
// thank https://github.com/brpaz/echozap
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			if err != nil {
				log = log.With(zap.Error(err))
				c.Error(err)
			}

			// skip heartbeat request
			if c.Path() == "/status" {
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
