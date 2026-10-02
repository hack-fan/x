package xecho

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapLoggerWithSkipper(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(zap.NewNop())
	e.Use(ZapLoggerWithSkipper(zap.New(core), NewSkipper([]SkipRule{
		{Method: http.MethodGet, Path: "/status", StatusCode: http.StatusOK},
	})))
	e.GET("/status", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	e.GET("/ok", func(c *echo.Context) error { return c.String(http.StatusOK, "hello") })
	e.GET("/missing", func(c *echo.Context) error { return echo.ErrNotFound })

	serve := func(path, requestID string) {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if requestID != "" {
			req.Header.Set(echo.HeaderXRequestID, requestID)
		}
		e.ServeHTTP(httptest.NewRecorder(), req)
	}

	serve("/status", "")
	assert.Equal(t, 0, logs.Len(), "skipped request should not be logged")

	serve("/ok", "req-1")
	entry := logs.TakeAll()[0]
	assert.Equal(t, "Success", entry.Message)
	fields := entry.ContextMap()
	assert.EqualValues(t, http.StatusOK, fields["status"])
	assert.EqualValues(t, 5, fields["size"])
	assert.Equal(t, "req-1", fields["request_id"])

	serve("/missing", "")
	entry = logs.TakeAll()[0]
	assert.Equal(t, "Client Error", entry.Message)
	assert.EqualValues(t, http.StatusNotFound, entry.ContextMap()["status"])
	assert.NotContains(t, entry.ContextMap(), "request_id")
}

func TestZapLoggerServerErrorLoggedOnce(t *testing.T) {
	reqCore, reqLogs := observer.New(zap.InfoLevel)
	errCore, errLogs := observer.New(zap.ErrorLevel)
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(zap.New(errCore))
	e.Use(ZapLogger(zap.New(reqCore)))
	e.GET("/boom", func(c *echo.Context) error { return errors.New("boom") })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, 1, errLogs.Len(), "server error should be logged once")
	assert.Equal(t, "Server Error", reqLogs.TakeAll()[0].Message)
}
