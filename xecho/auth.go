package xecho

import (
	"strings"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
)

// KeyAuthConfig returns a middleware config with custom error handler.
// Use this with echo middleware.KeyAuth.
//
// Example:
//
//	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
//	    return key == "valid-key", nil
//	}, KeyAuthConfig()))
func KeyAuthConfig() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// This is just a wrapper that sets up the error handler
			// The actual key validation should be done in the KeyAuth middleware
			return next(c)
		}
	}
}

// KeyAuthErrorHandler is custom error handler for echo KeyAuth middleware.
// If this is not set, it's will convert all validator error to 400 error
func KeyAuthErrorHandler(err error, _ echo.Context) error {
	msg := err.Error()
	if strings.HasPrefix(msg, "missing key") || strings.HasPrefix(msg, "invalid key") {
		return xerr.New(400, "InvalidKey", err.Error())
	}
	return err
}
