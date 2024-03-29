package xecho

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
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
		} else if ve, ok := err.(validator.ValidationErrors); ok {
			resp = xerr.New(400, "BadRequest", ve.Error())
		} else if ee, ok := err.(*echo.HTTPError); ok {
			if he, ok := ee.Internal.(*xerr.Error); ok {
				resp = he
			} else {
				// echo errors
				resp = xerr.New(ee.Code, strings.ReplaceAll(http.StatusText(ee.Code), " ", ""),
					fmt.Sprintf("%v %s", ee.Message, ee.Unwrap()))
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// gorm not found
			resp = xerr.New(404, "NotFound", "record not found")
		} else {
			resp = xerr.New(500, "ServerError", err.Error())
		}

		// hide the server error message to client
		if resp.StatusCode() >= 500 {
			logger.Error(resp.Message)
			resp = xerr.ServerError
		}

		// echo need this
		if !c.Response().Committed {
			if c.Request().Method == echo.HEAD {
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
}
