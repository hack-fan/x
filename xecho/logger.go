package xecho

import (
	"slices"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

// Skipper decides whether a request should not be logged.
type Skipper = middleware.Skipper

// SkipRule must be fully equal
type SkipRule struct {
	Method     string
	Path       string
	StatusCode int
}

// NewSkipper gen a logger skipper
func NewSkipper(rules []SkipRule) Skipper {
	return func(c *echo.Context) bool {
		_, status := echo.ResolveResponseStatus(c.Response(), nil)
		return slices.Contains(rules, SkipRule{
			Method:     c.Request().Method,
			Path:       c.Path(),
			StatusCode: status,
		})
	}
}

// LoggerSkipper skip the heartbeat /status log
func LoggerSkipper(c *echo.Context) bool {
	return c.Path() == "/status"
}

// ZapLoggerWithSkipper use zap as request logger.
// Errors are handled by the echo error handler before logging, so the logged status is final.
// The skipper is checked after the request is handled, so it can use the response status.
// Register it as the first (outermost) middleware, because the response is sent once it handles the error.
func ZapLoggerWithSkipper(log *zap.Logger, skipper Skipper) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		HandleError:     true,
		LogLatency:      true,
		LogRemoteIP:     true,
		LogHost:         true,
		LogMethod:       true,
		LogURI:          true,
		LogStatus:       true,
		LogResponseSize: true,
		LogUserAgent:    true,
		LogRequestID:    true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if skipper != nil && skipper(c) {
				return nil
			}

			fields := []zap.Field{
				zap.String("remote_ip", v.RemoteIP),
				zap.String("time", v.Latency.String()),
				zap.String("host", v.Host),
				zap.String("request", v.Method+" "+v.URI),
				zap.Int("status", v.Status),
				zap.Int64("size", v.ResponseSize),
				zap.String("user_agent", v.UserAgent),
			}
			if v.RequestID != "" {
				fields = append(fields, zap.String("request_id", v.RequestID))
			}
			if v.Error != nil {
				fields = append(fields, zap.Error(v.Error))
			}

			// log all at info level
			// if there is a server error, echo error handler will log it as error there.
			switch {
			case v.Status >= 500:
				log.Info("Server Error", fields...)
			case v.Status >= 400:
				log.Info("Client Error", fields...)
			case v.Status >= 300:
				log.Info("Redirection", fields...)
			default:
				log.Info("Success", fields...)
			}
			return nil
		},
	})
}

// ZapLogger use zap for echo logger
func ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return ZapLoggerWithSkipper(log, nil)
}
