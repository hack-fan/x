package xecho

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LoggerSkipper skip the heartbeat /status log
func LoggerSkipper(c echo.Context) bool {
	return c.Path() == "/status"
}

// LoggerMid skip /status endpoint, will be deprecated.
func LoggerMid() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: LoggerSkipper,
	})
}
