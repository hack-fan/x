package xecho

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LoggerSkipper skip the heartbeat /status log
func LoggerSkipper(c echo.Context) bool {
	return c.Path() == "/status"
}

// Deprecated: LoggerMid skip /status endpoint, use RequestLoggerWithConfig directly.
func LoggerMid() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: LoggerSkipper,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			return nil
		},
	})
}
