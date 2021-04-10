package xecho

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// NewErrorHandler return a customize echo's HTTP error handler.
func NewErrorHandler(logger *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// the final response body
		var resp *xerr.Error

		if he, ok := err.(*xerr.Error); ok {
			// custom error by this package
			resp = he
		} else if ee, ok := err.(*echo.HTTPError); ok {
			// echo errors
			resp = xerr.New(ee.Code, strings.ReplaceAll(http.StatusText(ee.Code), " ", ""),
				fmt.Sprintf("%v", ee.Message))
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// gorm not found
			resp = xerr.New(404, "NotFound", "record not found")
		} else {
			// server errors output
			msg := ""
			if c.Echo().Debug {
				msg = err.Error()
			}
			resp = xerr.New(500, "ServerError", msg)
		}

		// echo need this
		if !c.Response().Committed {
			if c.Request().Method == echo.HEAD {
				err = c.NoContent(resp.StatusCode())
			} else {
				err = c.JSON(resp.StatusCode(), resp)
			}
			if err != nil {
				// log hook only show the message field, so write err as message
				logger.Error(err.Error())
			}
		}
	}
}
