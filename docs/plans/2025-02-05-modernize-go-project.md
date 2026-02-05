# Go Project Modernization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Modernize all packages in github.com/hack-fan/x to use Go 1.22+ best practices including generics, new standard library packages, improved error handling, and better code organization.

**Architecture:** Incremental modernization by package starting with core dependencies (xerr → xlog) and working outward to dependent packages. Each package is fully modernized and tested before moving to the next. Breaking changes are acceptable.

**Tech Stack:** Go 1.22+, generics (Go 1.18+), slices/cmaps packages (Go 1.21+), errors.Join (Go 1.20+), any instead of interface{}, context propagation

---

## Task 1: Modernize xerr Package

**Files:**
- Modify: `xerr/xerr.go`
- Modify: `xerr/go.mod`
- Test: `xerr/xerr_test.go` (if exists, otherwise create)

### Step 1: Add modern error helpers

Add these helper functions to `xerr/xerr.go` after line 134:

```go
// IsServerError check if error is a 5xx server error
func IsServerError(err error) bool {
	e, ok := As(err)
	if !ok {
		return false
	}
	return e.code >= 500 && e.code < 600
}

// Join wraps errors.Join to create a multi-error xerr
func Join(code int, key string, msg string, errs ...error) *Error {
	joined := errors.Join(errs...)
	return &Error{
		err:     joined,
		code:    code,
		Key:     key,
		Message: msg,
	}
}

// UnwrapAll returns all wrapped errors
func (e *Error) UnwrapAll() []error {
	var result []error
	err := error(e)
	for err != nil {
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
			if err != nil {
				result = append(result, err)
			}
		} else {
			break
		}
	}
	return result
}
```

### Step 2: Replace interface{} with any

In `xerr/xerr.go` line 33, replace:
```go
func Newf(code int, key string, format string, a ...interface{}) *Error {
```

With:
```go
func Newf(code int, key string, format string, a ...any) *Error {
```

### Step 3: Improve godoc documentation

Add examples to package documentation at the top of `xerr/xerr.go` after line 9:

```go
// Package xerr provides custom error handling with HTTP status codes.
//
// # Basic Usage
//
//	err := xerr.New(404, "NotFound", "User not found")
//	fmt.Println(err.StatusCode()) // 404
//
// # Error Checking
//
//	if xerr.Is(err, "NotFound") {
//	    // handle not found
//	}
//	if xerr.IsClientError(err) {
//	    // handle 4xx errors
//	}
//
// # Error Wrapping
//
//	err := xerr.Newf(500, "DatabaseError", "failed to connect: %w", dbErr)
//	if errors.Is(err, dbErr) {
//	    // database connection failed
//	}
package xerr
```

### Step 4: Create comprehensive tests

Create `xerr/xerr_test.go`:

```go
package xerr

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(404, "NotFound", "Resource not found")
	if err.StatusCode() != 404 {
		t.Errorf("expected 404, got %d", err.StatusCode())
	}
	if err.Key != "NotFound" {
		t.Errorf("expected NotFound, got %s", err.Key)
	}
	if err.Message != "Resource not found" {
		t.Errorf("expected 'Resource not found', got %s", err.Message)
	}
}

func TestIs(t *testing.T) {
	err := New(404, "NotFound", "not found")
	if !Is(err, "NotFound") {
		t.Error("Is should return true for matching key")
	}
	if Is(err, "OtherError") {
		t.Error("Is should return false for non-matching key")
	}
}

func TestIsCode(t *testing.T) {
	err := New(404, "NotFound", "not found")
	if !IsCode(err, 404) {
		t.Error("IsCode should return true for matching code")
	}
	if IsCode(err, 500) {
		t.Error("IsCode should return false for non-matching code")
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		want    bool
	}{
		{"400 error", New(400, "BadRequest", "bad"), true},
		{"404 error", New(404, "NotFound", "not found"), true},
		{"499 error", New(499, "ClientClosed", "closed"), true},
		{"500 error", New(500, "ServerError", "server error"), false},
		{"200 error", New(200, "OK", "ok"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsClientError(tt.err); got != tt.want {
				t.Errorf("IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		want    bool
	}{
		{"500 error", New(500, "ServerError", "server error"), true},
		{"503 error", New(503, "ServiceUnavailable", "unavailable"), true},
		{"404 error", New(404, "NotFound", "not found"), false},
		{"400 error", New(400, "BadRequest", "bad"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsServerError(tt.err); got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := Newf(500, "WrappedError", "wrapped: %w", baseErr)

	if !errors.Is(err, baseErr) {
		t.Error("Newf should wrap the base error")
	}
}

func TestJoin(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	err := Join(500, "MultiError", "multiple errors", err1, err2)

	if err.StatusCode() != 500 {
		t.Errorf("expected 500, got %d", err.StatusCode())
	}
	if err.Key != "MultiError" {
		t.Errorf("expected MultiError, got %s", err.Key)
	}
}

func TestAs(t *testing.T) {
	err := New(404, "NotFound", "not found")
	got, ok := As(err)
	if !ok {
		t.Fatal("As should return true for *Error type")
	}
	if got.Key != "NotFound" {
		t.Errorf("expected NotFound, got %s", got.Key)
	}

	// Test with non-xerr error
	_, ok = As(errors.New("standard error"))
	if ok {
		t.Error("As should return false for non-*Error type")
	}
}
```

### Step 5: Run tests to verify they pass

Run:
```bash
cd xerr && go test -v
```

Expected: All tests PASS

### Step 6: Run linting

Run:
```bash
cd xerr && gofmt -w . && go vet ./...
```

Expected: No errors, code formatted

### Step 7: Commit xerr changes

```bash
git add xerr/
git commit -m "refactor(xerr): modernize with Go 1.22+ patterns

- Add IsServerError() helper
- Add Join() for multi-error support
- Add UnwrapAll() method
- Replace interface{} with any
- Improve godoc documentation
- Add comprehensive test coverage
"
```

---

## Task 2: Modernize xlog Package

**Files:**
- Modify: `xlog/log.go`
- Modify: `xlog/wework.go`
- Test: `xlog/log_test.go` (create)

### Step 1: Add context support to HTTP client

In `xlog/log.go`, add a new context-aware HTTP client function after line 16:

```go
// httpPostWithContext sends HTTP POST with context and timeout
func httpPostWithContext(ctx context.Context, url string, body any) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return httpc.Do(req)
}
```

Add import for context at the top:
```go
import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
```

### Step 2: Add composite hooks implementation

After the Hook interface definition (after line 14), add:

```go
// Hooks is a slice of Hook that implements Hook interface
type Hooks []Hook

// Process calls each hook in sequence
func (h Hooks) Process(entry zapcore.Entry) error {
	for _, hook := range h {
		if err := hook.Process(entry); err != nil {
			return err
		}
	}
	return nil
}
```

### Step 3: Add logger builder with options

Add after the `New` function (after line 33):

```go
// Config is the logger configuration
type Config struct {
	Debug    bool
	Hooks    []Hook
	Level    zapcore.Level
	Encoding string // "json" or "console"
}

// Option configures the logger
type Option func(*Config)

// WithDebug sets debug mode
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithHooks adds hooks
func WithHooks(hooks ...Hook) Option {
	return func(c *Config) {
		c.Hooks = hooks
	}
}

// WithLevel sets the log level
func WithLevel(level zapcore.Level) Option {
	return func(c *Config) {
		c.Level = level
	}
}

// WithEncoding sets the encoding format
func WithEncoding(encoding string) Option {
	return func(c *Config) {
		c.Encoding = encoding
	}
}

// NewWithConfig creates a logger with configuration options
func NewWithConfig(opts ...Option) *zap.Logger {
	cfg := &Config{
		Debug:    false,
		Level:    zapcore.InfoLevel,
		Encoding: "json",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	var zapConfig zap.Config
	if cfg.Debug {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = cfg.Encoding
	}
	zapConfig.Level = zap.NewAtomicLevelAt(cfg.Level)

	log, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}

	for _, hook := range cfg.Hooks {
		log = log.WithOptions(zap.Hooks(hook.Process))
	}

	return log
}
```

### Step 4: Replace interface{} with any

In `xlog/log.go`, replace any remaining `interface{}` with `any`.

### Step 5: Create tests for xlog

Create `xlog/log_test.go`:

```go
package xlog

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

// mockHook is a test hook that records entries
type mockHook struct {
	entries []zapcore.Entry
}

func (m *mockHook) Process(entry zapcore.Entry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func TestNew(t *testing.T) {
	// Test production logger
	log := New(false)
	if log == nil {
		t.Error("New should return a non-nil logger")
	}

	// Test debug logger
	log = New(true)
	if log == nil {
		t.Error("New should return a non-nil logger in debug mode")
	}
}

func TestNewWithHooks(t *testing.T) {
	hook := &mockHook{}
	log := New(false, hook)

	if log == nil {
		t.Fatal("New should return a non-nil logger")
	}

	// Log something
	log.Info("test message")

	// Wait for async processing
	log.Sync()

	if len(hook.entries) == 0 {
		t.Error("hook should have received log entries")
	}
}

func TestHooks_Process(t *testing.T) {
	hook1 := &mockHook{}
	hook2 := &mockHook{}
	hooks := Hooks{hook1, hook2}

	entry := zapcore.Entry{Message: "test"}
	err := hooks.Process(entry)

	if err != nil {
		t.Errorf("Process should not return error, got %v", err)
	}

	if len(hook1.entries) != 1 {
		t.Errorf("hook1 should have 1 entry, got %d", len(hook1.entries))
	}
	if len(hook2.entries) != 1 {
		t.Errorf("hook2 should have 1 entry, got %d", len(hook2.entries))
	}
}

func TestNewWithConfig(t *testing.T) {
	hook := &mockHook{}
	log := NewWithConfig(
		WithDebug(true),
		WithHooks(hook),
		WithLevel(zapcore.DebugLevel),
		WithEncoding("console"),
	)

	if log == nil {
		t.Fatal("NewWithConfig should return a non-nil logger")
	}

	log.Debug("debug message")
	log.Sync()

	if len(hook.entries) == 0 {
		t.Error("hook should have received log entries")
	}
}

func TestWithDebug(t *testing.T) {
	opt := WithDebug(true)
	cfg := &Config{}
	opt(cfg)

	if !cfg.Debug {
		t.Error("WithDebug should set Debug to true")
	}
}

func TestWithLevel(t *testing.T) {
	opt := WithLevel(zapcore.ErrorLevel)
	cfg := &Config{}
	opt(cfg)

	if cfg.Level != zapcore.ErrorLevel {
		t.Errorf("WithLevel should set Level to ErrorLevel, got %v", cfg.Level)
	}
}

func TestWithEncoding(t *testing.T) {
	opt := WithEncoding("console")
	cfg := &Config{}
	opt(cfg)

	if cfg.Encoding != "console" {
		t.Errorf("WithEncoding should set Encoding to console, got %s", cfg.Encoding)
	}
}
```

### Step 6: Run tests

Run:
```bash
cd xlog && go test -v
```

Expected: All tests PASS

### Step 7: Run linting

Run:
```bash
cd xlog && gofmt -w . && go vet ./...
```

Expected: No errors

### Step 8: Commit xlog changes

```bash
git add xlog/
git commit -m "refactor(xlog): modernize with context support and options pattern

- Add context-aware HTTP client
- Add Hooks composite type for multiple hooks
- Add NewWithConfig with functional options
- Replace interface{} with any
- Add comprehensive test coverage
"
```

---

## Task 3: Modernize xtype Package

**Files:**
- Modify: `xtype/json.go`
- Modify: `xtype/values.go`
- Test: `xtype/json_test.go`, `xtype/values_test.go`

### Step 1: Add generic JSON helpers

Add to `xtype/json.go`:

```go
package xtype

import (
	"encoding/json"
	"fmt"
)

// ToJSON converts any value to JSON string
func ToJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

// ToJSONPretty converts any value to pretty-printed JSON string
func ToJSONPretty(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// FromJSON parses JSON string into target
func FromJSON(data string, target any) error {
	return json.Unmarshal([]byte(data), target)
}

// FromJSONMust parses JSON string into target, panics on error
func FromJSONMust(data string, target any) {
	if err := FromJSON(data, target); err != nil {
		panic(err)
	}
}

// ToMap converts struct to map[string]any
func ToMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}
```

### Step 2: Add generic value helpers

Add to `xtype/values.go`:

```go
package xtype

import "cmp"

// Compare returns -1, 0, or 1 depending on ordering of a and b
func Compare[T cmp.Ordered](a, b T) int {
	return cmp.Compare(a, b)
}

// Min returns the minimum of two ordered values
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two ordered values
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp returns value clamped between min and max
func Clamp[T cmp.Ordered](v, min, max T) T {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Coalesce returns first non-zero value
func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// Ptr returns pointer to value
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns value pointed to, or default value if nil
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns value pointed to, or the provided default
func DerefOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
```

### Step 3: Create tests

Create `xtype/json_test.go`:

```go
package xtype

import (
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestToJSON(t *testing.T) {
	s := TestStruct{Name: "test", Age: 42}
	json := ToJSON(s)
	if json == "" {
		t.Error("ToJSON should not return empty string")
	}
	expected := `{"name":"test","age":42}`
	if json != expected {
		t.Errorf("expected %s, got %s", expected, json)
	}
}

func TestToJSONPretty(t *testing.T) {
	s := TestStruct{Name: "test", Age: 42}
	json := ToJSONPretty(s)
	if json == "" {
		t.Error("ToJSONPretty should not return empty string")
	}
}

func TestFromJSON(t *testing.T) {
	data := `{"name":"test","age":42}`
	var s TestStruct
	err := FromJSON(data, &s)
	if err != nil {
		t.Errorf("FromJSON should not error, got %v", err)
	}
	if s.Name != "test" || s.Age != 42 {
		t.Errorf("expected name=test, age=42, got name=%s, age=%d", s.Name, s.Age)
	}
}

func TestFromJSONMust(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// Expected to panic with invalid JSON
		}
	}()

	data := `{"name":"test","age":42}`
	var s TestStruct
	FromJSONMust(data, &s)
	if s.Name != "test" {
		t.Errorf("expected name=test, got %s", s.Name)
	}

	// This will panic (we catch it above)
	FromJSONMust("invalid", &s)
}

func TestToMap(t *testing.T) {
	s := TestStruct{Name: "test", Age: 42}
	m, err := ToMap(s)
	if err != nil {
		t.Errorf("ToMap should not error, got %v", err)
	}
	if m["name"] != "test" {
		t.Errorf("expected name=test, got %v", m["name"])
	}
	if m["age"].(float64) != float64(42) {
		t.Errorf("expected age=42, got %v", m["age"])
	}
}
```

Create `xtype/values_test.go`:

```go
package xtype

import (
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"less", 1, 2, -1},
		{"equal", 2, 2, 0},
		{"greater", 3, 2, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compare(tt.a, tt.b); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMin(t *testing.T) {
	if got := Min(3, 5); got != 3 {
		t.Errorf("Min() = %v, want 3", got)
	}
	if got := Min(5, 3); got != 3 {
		t.Errorf("Min() = %v, want 3", got)
	}
}

func TestMax(t *testing.T) {
	if got := Max(3, 5); got != 5 {
		t.Errorf("Max() = %v, want 5", got)
	}
	if got := Max(5, 3); got != 5 {
		t.Errorf("Max() = %v, want 5", got)
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want int
	}{
		{5, 1, 10, 5},
		{0, 1, 10, 1},
		{15, 1, 10, 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := Clamp(tt.v, tt.min, tt.max); got != tt.want {
				t.Errorf("Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoalesce(t *testing.T) {
	if got := Coalesce(0, 1, 2); got != 1 {
		t.Errorf("Coalesce() = %v, want 1", got)
	}
	if got := Coalesce("", "a", "b"); got != "a" {
		t.Errorf("Coalesce() = %v, want 'a'", got)
	}
}

func TestPtr(t *testing.T) {
	p := Ptr(42)
	if p == nil {
		t.Fatal("Ptr() should not return nil")
	}
	if *p != 42 {
		t.Errorf("Ptr() = %v, want 42", *p)
	}
}

func TestDeref(t *testing.T) {
	v := 42
	if got := Deref(&v); got != 42 {
		t.Errorf("Deref() = %v, want 42", got)
	}
	if got := Deref[*int](nil); got != 0 {
		t.Errorf("Deref(nil) = %v, want 0", got)
	}
}

func TestDerefOr(t *testing.T) {
	v := 42
	if got := DerefOr(&v, 99); got != 42 {
		t.Errorf("DerefOr() = %v, want 42", got)
	}
	if got := DerefOr(nil, 99); got != 99 {
		t.Errorf("DerefOr(nil, 99) = %v, want 99", got)
	}
}
```

### Step 4: Run tests

```bash
cd xtype && go test -v
```

Expected: All tests PASS

### Step 5: Commit xtype changes

```bash
git add xtype/
git commit -m "refactor(xtype): add generic helpers using Go 1.18+ features

- Add generic JSON conversion functions
- Add generic comparison functions using cmp package
- Add generic min/max/clamp functions
- Add generic pointer helpers (Ptr, Deref, DerefOr)
- Replace all interface{} with any
- Add comprehensive tests
"
```

---

## Task 4: Modernize xecho Package

**Files:**
- Modify: `xecho/auth.go`
- Modify: `xecho/error.go`
- Modify: `xecho/logger.go`
- Modify: `xecho/page.go`
- Test: `xecho/*_test.go`

### Step 1: Replace interface{} with any

In all files under xecho, replace `interface{}` with `any`.

### Step 2: Improve error handler

Modernize `xecho/error.go` with better context support:

```go
package xecho

import (
	"net/http"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
)

// ErrorHandler converts xerr to HTTP responses
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var code int
		var response any

		if xe := xerr.AsXErr(err); xe != nil {
			code = xe.StatusCode()
			response = map[string]any{
				"error":   xe.Key,
				"message": xe.Message,
			}
		} else {
			code = http.StatusInternalServerError
			response = map[string]string{
				"error":   "InternalServerError",
				"message": http.StatusText(code),
			}
		}

		// Log the error
		c.Logger().Error(err.Error())

		// Send JSON response
		if !c.Response().Committed {
			c.JSON(code, response)
		}
	}
}

// Add convenience method to xerr package first:
// In xerr/xerr.go, add:
// func AsXErr(err error) *Error {
//     e, _ := As(err)
//     return e
// }
```

### Step 3: Add generic context helpers

Add to `xecho/auth.go`:

```go
package xecho

import (
	"github.com/labstack/echo/v4"
)

// GetCtx retrieves a context value with type safety
func GetCtx[T any](c echo.Context, key string) (T, bool) {
	var zero T
	v := c.Get(key)
	if v == nil {
		return zero, false
	}
	tv, ok := v.(T)
	return tv, ok
}

// SetCtx stores a context value
func SetCtx[T any](c echo.Context, key string, val T) {
	c.Set(key, val)
}
```

### Step 4: Update auth middleware

Improve `xecho/auth.go` KeyAuthErrorHandler:

```go
package xecho

import (
	"strings"

	"github.com/hack-fan/x/xerr"
	"github.com/labstack/echo/v4"
)

// KeyAuthErrorHandler is custom error handler for echo KeyAuth middleware
func KeyAuthErrorHandler(err error, c echo.Context) error {
	msg := err.Error()
	if strings.HasPrefix(msg, "missing key") || strings.HasPrefix(msg, "invalid key") {
		return xerr.New(400, "InvalidKey", msg)
	}
	return err
}

// KeyAuthConfig returns a preconfigured KeyAuthConfig
func KeyAuthConfig(validator func(string, echo.Context) (bool, error)) echo.KeyAuthConfig {
	return echo.KeyAuthConfig{
		KeyLookup:   "header:Authorization",
		Validator:   validator,
		ErrorHandler: KeyAuthErrorHandler,
	}
}
```

### Step 5: Create tests

Create `xecho/auth_test.go`:

```go
package xecho

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGetCtx(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test not found
	_, ok := GetCtx[string](c, "missing")
	if ok {
		t.Error("GetCtx should return false for missing key")
	}

	// Test found
	SetCtx(c, "key", "value")
	val, ok := GetCtx[string](c, "key")
	if !ok {
		t.Error("GetCtx should return true for existing key")
	}
	if val != "value" {
		t.Errorf("GetCtx = %v, want 'value'", val)
	}

	// Test type mismatch
	SetCtx(c, "num", 42)
	_, ok = GetCtx[string](c, "num")
	if ok {
		t.Error("GetCtx should return false for type mismatch")
	}
}

func TestSetCtx(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	SetCtx(c, "key", "value")
	val, ok := GetCtx[string](c, "key")
	if !ok || val != "value" {
		t.Errorf("SetCtx/GetCtx failed, got %v, %v", val, ok)
	}
}
```

### Step 6: Run tests

```bash
cd xecho && go test -v
```

### Step 7: Commit xecho changes

```bash
git add xecho/
git commit -m "refactor(xecho): add generic context helpers

- Add GetCtx/SetCtx generic functions for type-safe context values
- Add KeyAuthConfig helper
- Replace interface{} with any
- Improve error handler with proper response format
- Add tests for context helpers
"
```

---

## Task 5: Modernize xdb Package

**Files:**
- Modify: `xdb/xdb.go`

### Step 1: Add context support throughout

Replace any existing DB functions to accept context as first parameter.

### Step 2: Replace interface{} with any

### Step 3: Commit

```bash
git add xdb/
git commit -m "refactor(xdb): add context support and modernize

- Add context.Context as first parameter to DB operations
- Replace interface{} with any
"
```

---

## Task 6: Modernize rdb Package

**Files:**
- Modify: `rdb/redis.go`

### Step 1: Add context support

Ensure all Redis operations accept context.

### Step 2: Replace interface{} with any

### Step 3: Commit

```bash
git add rdb/
git commit -m "refactor(rdb): add context support

- Add context.Context to Redis operations
- Replace interface{} with any
"
```

---

## Task 7: Modernize xmp Package

**Files:**
- Modify: `xmp/mp.go`
- Modify: `xmp/token.go`

### Step 1: Add context support

### Step 2: Replace interface{} with any

### Step 3: Add structured logging integration

### Step 4: Commit

```bash
git add xmp/
git commit -m "refactor(xmp): modernize WeChat Mini Program utilities

- Add context support throughout
- Replace interface{} with any
- Improve error handling
"
```

---

## Task 8: Modernize xpay Package

**Files:**
- Modify: `xpay/pay.go`

### Step 1: Add context support

### Step 2: Replace interface{} with any

### Step 3: Commit

```bash
git add xpay/
git commit -m "refactor(xpay): modernize payment utilities

- Add context support
- Replace interface{} with any
- Improve error handling
"
```

---

## Task 9: Modernize xobj Package

**Files:**
- Modify: `xobj/cos.go`
- Modify: `xobj/obj.go`

### Step 1: Add context support

### Step 2: Replace interface{} with any

### Step 3: Commit

```bash
git add xobj/
git commit -m "refactor(xobj): modernize object storage

- Add context support to COS operations
- Replace interface{} with any
- Improve error handling
"
```

---

## Task 10: Final Validation

### Step 1: Run all tests

```bash
make test
```

Expected: All tests PASS across all packages

### Step 2: Run linter

```bash
make lint
```

Expected: No errors

### Step 3: Update go.mod if needed

```bash
go mod tidy
```

### Step 4: Final commit

```bash
git add .
git commit -m "chore: update dependencies after modernization"
```

### Step 5: Verify build

```bash
go build ./...
```

Expected: All packages build successfully

---

## Completion Checklist

- [ ] All packages modernized with generics where applicable
- [ ] All `interface{}` replaced with `any`
- [ ] Context support added throughout
- [ ] Comprehensive tests added
- [ ] All tests passing
- [ ] Code formatted with gofmt
- [ ] go vet shows no errors
- [ ] Documentation updated
- [ ] All changes committed

---

## Notes for Implementation

- **DRY:** Common patterns extracted into reusable helpers
- **YAGNI:** No unnecessary abstractions, only what's needed
- **TDD:** Tests written before or alongside implementation
- **Frequent commits:** Each small change is committed
- **Breaking changes acceptable:** We're not maintaining backward compatibility
