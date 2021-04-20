package xerr

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// ServerError always the same
var ServerError = New(500, "ServerError",
	"There was a little problem on the server side, please report it to us or try again later.")

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

// ParseResp can parse http response, if there is a error, it would read and close the body for error messages.
func ParseResp(resp *http.Response) *Error {
	if resp == nil {
		return &Error{
			code:    500,
			Key:     "ServerError",
			Message: "xerr can not parse a nil http response",
		}
	}
	if resp.StatusCode < 400 {
		return nil
	}
	var msg string
	defer resp.Body.Close()
	if strings.HasPrefix(resp.Header.Get("Content-Length"), "text") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return &Error{
				code:    500,
				Key:     "ServerError",
				Message: fmt.Sprintf("xerr parse response failed: %s", err),
			}
		}
		msg = string(body)
		if len(body) > 1000 {
			msg = msg[:1000]
		}
	}
	return &Error{
		code:    resp.StatusCode,
		Key:     strings.ReplaceAll(http.StatusText(resp.StatusCode), " ", ""),
		Message: msg,
	}
}

// Error makes it compatible with `error` interface.
func (e *Error) Error() string {
	return e.Message
}

// StatusCode is http status code
func (e *Error) StatusCode() int {
	return e.code
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
