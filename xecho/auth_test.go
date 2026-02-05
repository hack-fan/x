package xecho

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestKeyAuthErrorHandler(t *testing.T) {
	tests := []struct {
		name           string
		inputErr       error
		expectedStatus int
		expectXErr     bool
	}{
		{
			name:           "missing key error",
			inputErr:       errors.New("missing key in request"),
			expectedStatus: 400,
			expectXErr:     true,
		},
		{
			name:           "invalid key error",
			inputErr:       errors.New("invalid key format"),
			expectedStatus: 400,
			expectXErr:     true,
		},
		{
			name:           "other error passthrough",
			inputErr:       errors.New("some other error"),
			expectedStatus: 0,
			expectXErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			c := e.NewContext(nil, nil)

			result := KeyAuthErrorHandler(tt.inputErr, c)

			if tt.expectXErr {
				xErr, ok := result.(*xerr.Error)
				assert.True(t, ok, "expected xerr.Error")
				assert.Equal(t, tt.expectedStatus, xErr.StatusCode())
			}
		})
	}
}

func TestGetCtx(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Run("get non-existent key", func(t *testing.T) {
		val, ok := GetCtx[string](c, "nonexistent")
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("get existing key", func(t *testing.T) {
		SetCtx(c, "test_key", "test_value")

		val, ok := GetCtx[string](c, "test_key")
		assert.True(t, ok)
		assert.Equal(t, "test_value", val)
	})

	t.Run("type mismatch", func(t *testing.T) {
		SetCtx(c, "number", 123)

		val, ok := GetCtx[string](c, "number")
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestSetCtx(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Run("set string value", func(t *testing.T) {
		SetCtx(c, "user_id", "12345")

		val, ok := GetCtx[string](c, "user_id")
		assert.True(t, ok)
		assert.Equal(t, "12345", val)
	})

	t.Run("set int value", func(t *testing.T) {
		SetCtx(c, "count", 42)

		val, ok := GetCtx[int](c, "count")
		assert.True(t, ok)
		assert.Equal(t, 42, val)
	})

	t.Run("set bool value", func(t *testing.T) {
		SetCtx(c, "is_admin", true)

		val, ok := GetCtx[bool](c, "is_admin")
		assert.True(t, ok)
		assert.True(t, val)
	})

	t.Run("overwrite existing value", func(t *testing.T) {
		SetCtx(c, "key", "old_value")
		SetCtx(c, "key", "new_value")

		val, ok := GetCtx[string](c, "key")
		assert.True(t, ok)
		assert.Equal(t, "new_value", val)
	})
}

func TestMustGetCtx(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	t.Run("get existing key", func(t *testing.T) {
		SetCtx(c, "required_key", "required_value")

		val := MustGetCtx[string](c, "required_key")
		assert.Equal(t, "required_value", val)
	})

	t.Run("panic on non-existent key", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetCtx[string](c, "nonexistent_key")
		})
	})
}

func TestContextIntegration(t *testing.T) {
	// Integration test showing typical usage pattern
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Simulate middleware setting user context
	SetCtx(c, "user_id", "user123")
	SetCtx(c, "is_admin", false)
	SetCtx(c, "request_count", 5)

	// Handler retrieves context
	userID, ok := GetCtx[string](c, "user_id")
	assert.True(t, ok)
	assert.Equal(t, "user123", userID)

	isAdmin := MustGetCtx[bool](c, "is_admin")
	assert.False(t, isAdmin)

	count := MustGetCtx[int](c, "request_count")
	assert.Equal(t, 5, count)
}
