package xecho

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// LoggerSkipper skip the heartbeat /status log
func LoggerSkipper(c echo.Context) bool {
	if c.Path() == "/status" {
		return true
	}
	return false
}

func LoggerMid() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: LoggerSkipper,
	})
}
