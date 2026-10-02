package xecho

import (
	"cmp"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewErrorHandler returns a customized echo's HTTP error handler.
func NewErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		// the response is sent, e.g. the error was already handled by the request logger
		if r, uErr := echo.UnwrapResponse(c.Response()); uErr == nil && r.Committed {
			return
		}

		// the final response body
		resp := toXErr(err)

		// hide the server error message to client
		if resp.StatusCode() >= 500 {
			logger.Error(resp.Message)
			resp = xerr.ServerError
		}

		if c.Request().Method == http.MethodHead {
			err = c.NoContent(resp.StatusCode())
		} else {
			err = c.JSON(resp.StatusCode(), resp)
		}
		if err != nil {
			// log the resp sent error
			logger.Error(err.Error())
		}
	}
}

// toXErr converts any error returned by handlers to the response body.
func toXErr(err error) *xerr.Error {
	if e, ok := xerr.As(err); ok {
		// custom error by xerr, use it directly
		return e
	}
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		return xerr.New(400, "BadRequest", ve.Error())
	}
	// echo errors, including sentinels like echo.ErrNotFound which are not *echo.HTTPError
	if code := echo.StatusCode(err); code != 0 {
		msg := err.Error()
		if he, ok := errors.AsType[*echo.HTTPError](err); ok {
			msg = cmp.Or(he.Message, http.StatusText(code))
			if inner := he.Unwrap(); inner != nil {
				msg += ": " + inner.Error()
			}
		}
		return xerr.New(code, strings.ReplaceAll(http.StatusText(code), " ", ""), msg)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return xerr.New(404, "NotFound", "record not found")
	}
	return xerr.New(500, "ServerError", err.Error())
}

// ErrorHandler is a convenience function that returns an error handler using a default logger.
// Use this for quick setup without custom logger configuration.
func ErrorHandler() echo.HTTPErrorHandler {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return NewErrorHandler(logger)
}
