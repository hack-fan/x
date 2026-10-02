package xecho

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TestNewErrorHandler(t *testing.T) {
	validationErr := validator.New().Struct(struct {
		Name string `validate:"required"`
	}{})

	tests := []struct {
		name       string
		err        error
		method     string
		wantStatus int
		wantKey    string
	}{
		{name: "xerr", err: fmt.Errorf("wrap: %w", xerr.New(403, "Forbidden", "no")), wantStatus: 403, wantKey: "Forbidden"},
		{name: "validation", err: validationErr, wantStatus: 400, wantKey: "BadRequest"},
		{name: "echo http error", err: echo.NewHTTPError(http.StatusConflict, "dup"), wantStatus: 409, wantKey: "Conflict"},
		{name: "echo sentinel error", err: echo.ErrNotFound, wantStatus: 404, wantKey: "NotFound"},
		{name: "gorm not found", err: gorm.ErrRecordNotFound, wantStatus: 404, wantKey: "NotFound"},
		{name: "unknown error is hidden", err: errors.New("db password leaked"), wantStatus: 500, wantKey: "ServerError"},
		{name: "head request has no body", err: echo.ErrNotFound, method: http.MethodHead, wantStatus: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(cmp.Or(tt.method, http.MethodGet), "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			NewErrorHandler(zap.NewNop())(c, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantKey == "" {
				assert.Empty(t, rec.Body.String())
				return
			}
			assert.Contains(t, rec.Body.String(), `"error":"`+tt.wantKey+`"`)
			assert.NotContains(t, rec.Body.String(), "leaked")
		})
	}
}

func TestNewErrorHandlerCommitted(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	assert.NoError(t, c.String(http.StatusOK, "done"))

	NewErrorHandler(zap.NewNop())(c, errors.New("late error"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "done", rec.Body.String())
}
