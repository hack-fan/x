package xlog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// mockHook is a test implementation of Hook interface
type mockHook struct {
	called      bool
	processFunc func(zapcore.Entry) error
}

func (m *mockHook) Process(entry zapcore.Entry) error {
	m.called = true
	if m.processFunc != nil {
		return m.processFunc(entry)
	}
	return nil
}

// TestNew tests basic logger creation
func TestNew(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{"production logger", false},
		{"debug logger", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.debug)
			defer log.Sync()

			if log == nil {
				t.Fatal("Expected non-nil logger")
			}

			// Test that logger works
			log.Info("test message")
		})
	}
}

// TestNewWithHooks tests logger creation with hooks
func TestNewWithHooks(t *testing.T) {
	hook := &mockHook{}
	log := New(false, hook)
	defer log.Sync()

	if log == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Trigger hook by logging
	log.Info("test message")

	if !hook.called {
		t.Error("Expected hook to be called")
	}
}

// TestHooks_Process tests the Hooks composite type
func TestHooks_Process(t *testing.T) {
	tests := []struct {
		name      string
		hooks     Hooks
		wantError bool
	}{
		{
			name:      "single hook",
			hooks:     Hooks{&mockHook{}},
			wantError: false,
		},
		{
			name: "multiple hooks",
			hooks: Hooks{
				&mockHook{},
				&mockHook{},
			},
			wantError: false,
		},
		{
			name: "hook returns error",
			hooks: Hooks{
				&mockHook{},
				&mockHook{processFunc: func(zapcore.Entry) error {
					return assertError("test error")
				}},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := zapcore.Entry{Message: "test"}
			err := tt.hooks.Process(entry)

			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestNewWithConfig tests logger creation with configuration options
func TestNewWithConfig(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		check   func(*testing.T, *zap.Logger)
	}{
		{
			name:    "default config",
			options: []Option{},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
			},
		},
		{
			name:    "with debug",
			options: []Option{WithDebug(true)},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
				log.Info("test")
			},
		},
		{
			name:    "with hooks",
			options: []Option{WithHooks(&mockHook{})},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
				log.Info("test")
			},
		},
		{
			name:    "with level",
			options: []Option{WithLevel(zapcore.DebugLevel)},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
				log.Debug("test")
			},
		},
		{
			name:    "with encoding",
			options: []Option{WithEncoding("console")},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
				log.Info("test")
			},
		},
		{
			name: "multiple options",
			options: []Option{
				WithDebug(true),
				WithLevel(zapcore.WarnLevel),
				WithEncoding("console"),
			},
			check: func(t *testing.T, log *zap.Logger) {
				if log == nil {
					t.Fatal("Expected non-nil logger")
				}
				log.Warn("test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := NewWithConfig(tt.options...)
			defer log.Sync()
			tt.check(t, log)
		})
	}
}

// TestWithDebug tests the WithDebug option
func TestWithDebug(t *testing.T) {
	cfg := &Config{}
	WithDebug(true)(cfg)

	if !cfg.Debug {
		t.Error("Expected Debug to be true")
	}
}

// TestWithLevel tests the WithLevel option
func TestWithLevel(t *testing.T) {
	cfg := &Config{}
	WithLevel(zapcore.ErrorLevel)(cfg)

	if cfg.Level != zapcore.ErrorLevel {
		t.Errorf("Expected Level to be ErrorLevel, got %v", cfg.Level)
	}
}

// TestWithEncoding tests the WithEncoding option
func TestWithEncoding(t *testing.T) {
	cfg := &Config{}
	WithEncoding("json")(cfg)

	if cfg.Encoding != "json" {
		t.Errorf("Expected Encoding to be json, got %s", cfg.Encoding)
	}
}

// TestHttpPostWithContext tests the context-aware HTTP client
func TestHttpPostWithContext(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		setupServer func() *httptest.Server
		wantError   bool
	}{
		{
			name: "successful request",
			ctx:  context.Background(),
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != "POST" {
						t.Errorf("Expected POST request, got %s", r.Method)
					}
					if r.Header.Get("Content-Type") != "application/json" {
						t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
					}
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantError: false,
		},
		{
			name: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx
			}(),
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantError: true,
		},
		{
			name: "context timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				defer cancel()
				return ctx
			}(),
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(time.Millisecond)
					w.WriteHeader(http.StatusOK)
				}))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			resp, err := httpPostWithContext(tt.ctx, server.URL, nil)

			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if resp == nil {
					t.Error("Expected non-nil response")
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}
}

// assertError is a test error type
type assertError string

func (e assertError) Error() string {
	return string(e)
}
