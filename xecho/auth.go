package xecho

import "github.com/labstack/echo/v4"

// KeyAuthErrorHandler is custom error handler for echo KeyAuth middleware.
// If this is not set, it's will convert all validator error to 401 error
func KeyAuthErrorHandler(err error, _ echo.Context) error {
	return err
}
