package xerr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Error custom struct
type Error struct {
	code    int
	Key     string `json:"error"`
	Message string `json:"message"`
}

// New Error
func New(code int, key string, msg string) *Error {
	return &Error{
		code:    code,
		Key:     key,
		Message: msg,
	}
}

// Newf create a Error use format
func Newf(code int, key string, format string, a ...interface{}) *Error {
	return &Error{
		code:    code,
		Key:     key,
		Message: fmt.Sprintf(format, a...),
	}
}

// Error makes it compatible with `error` interface.
func (e *Error) Error() string {
	return e.Key + ": " + e.Message
}

// ErrorHandler customize echo's HTTP error handler.
func ErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		key  = "ServerError"
		msg  string
	)
	c.Logger().Errorf("error in echo handler: %s", err)

	if he, ok := err.(*Error); ok {
		// custom error by this package
		code = he.code
		key = he.Key
		msg = he.Message
	} else if ee, ok := err.(*echo.HTTPError); ok {
		// echo errors
		code = ee.Code
		key = http.StatusText(code)
		msg = fmt.Sprintf("%v", ee.Message)
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// gorm not found
		code = http.StatusNotFound
		key = "NotFound"
		msg = http.StatusText(404)
	} else if c.Echo().Debug {
		// server errors debug
		msg = err.Error()
	} else {
		// server errors output
		msg = http.StatusText(code)
	}

	// echo need this
	if !c.Response().Committed {
		if c.Request().Method == echo.HEAD {
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, New(code, key, msg))
		}
		if err != nil {
			c.Logger().Error(err.Error())
		}
	}
}

// Is err the instance of Error,and has <key>?
func Is(err error, key string) bool {
	src, ok := err.(*Error)
	if !ok {
		return false
	}
	if src.Key == key {
		return true
	}
	return false
}

// IsCode check if the status code is <code>
func IsCode(err error, code int) bool {
	src, ok := err.(*Error)
	if !ok {
		return false
	}
	if src.code == code {
		return true
	}
	return false
}
