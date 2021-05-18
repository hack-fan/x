package xecho

import (
	"strings"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
)

// KeyAuthErrorHandler is custom error handler for echo KeyAuth middleware.
// If this is not set, it's will convert all validator error to 401 error
func KeyAuthErrorHandler(err error, _ echo.Context) error {
	msg := err.Error()
	if strings.HasPrefix(msg, "missing key") || strings.HasPrefix(msg, "invalid key") {
		return xerr.New(400, "InvalidKey", err.Error())
	}
	return err
}
