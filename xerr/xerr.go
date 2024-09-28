package xerr

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ServerError always the same
var ServerError = New(500, "ServerError",
	"There was an issue on the server side. Please report to us or try again later.")

// Error custom struct
type Error struct {
	err     error // support the Unwrap interface
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

// Newf create an Error use format
func Newf(code int, key string, format string, a ...interface{}) *Error {
	err := fmt.Errorf(format, a...)
	return &Error{
		err:     err,
		code:    code,
		Key:     key,
		Message: err.Error(),
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

// Unwrap support the Unwrap interface
func (e *Error) Unwrap() error {
	return e.err
}

// Is err the instance of Error,and has <key>?
func Is(err error, key string) bool {
	src, ok := As(err)
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
	src, ok := As(err)
	if !ok {
		return false
	}
	if src.code == code {
		return true
	}
	return false
}

func As(err error) (*Error, bool) {
	e := new(Error)
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func IsClientError(err error) bool {
	e, ok := As(err)
	if !ok {
		return false
	}
	if e.code >= 400 && e.code < 500 {
		return true
	}
	return false
}
