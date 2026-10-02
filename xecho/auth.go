package xecho

import (
	"errors"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// KeyAuthConfig returns an echo KeyAuth middleware config that validates keys
// with validator and reports missing or invalid keys by KeyAuthErrorHandler.
//
// Example:
//
//	e.Use(middleware.KeyAuthWithConfig(xecho.KeyAuthConfig(
//	    func(c *echo.Context, key string, _ middleware.ExtractorSource) (bool, error) {
//	        return subtle.ConstantTimeCompare([]byte(key), []byte("valid-key")) == 1, nil
//	    })))
func KeyAuthConfig(validator middleware.KeyAuthValidator) middleware.KeyAuthConfig {
	return middleware.KeyAuthConfig{
		Validator:    validator,
		ErrorHandler: KeyAuthErrorHandler,
	}
}

// KeyAuthErrorHandler is custom error handler for echo KeyAuth middleware.
// It converts missing or invalid key errors to a 400 xerr.Error,
// other errors (e.g. returned by the validator) are passed through.
func KeyAuthErrorHandler(_ *echo.Context, err error) error {
	if _, ok := errors.AsType[*middleware.ValueExtractorError](err); ok || errors.Is(err, middleware.ErrInvalidKey) {
		return xerr.New(400, "InvalidKey", err.Error())
	}
	return err
}
