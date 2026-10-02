package xlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
			defer func() { _ = log.Sync() }()

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
	defer func() { _ = log.Sync() }()

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
			defer func() { _ = log.Sync() }()
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

// TestWeworkSender_SendRobotMsg tests sending a robot message to the WeWork webhook.
func TestWeworkSender_SendRobotMsg(t *testing.T) {
	var got RobotMsg
	var gotKey, gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.URL.Query().Get("key")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if gotKey == "bad" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()
	s := WeworkSender{BaseURL: server.URL + "?key="}

	t.Run("empty key is a no-op", func(t *testing.T) {
		if err := s.SendRobotMsg("", "hello"); err != nil {
			t.Errorf("SendRobotMsg with empty key = %v, want nil", err)
		}
	})

	t.Run("send text message", func(t *testing.T) {
		if err := s.SendRobotMsg("k1", "hello"); err != nil {
			t.Fatalf("SendRobotMsg = %v", err)
		}
		if gotKey != "k1" || gotContentType != "application/json" {
			t.Errorf("key = %q, content type = %q", gotKey, gotContentType)
		}
		if got.MsgType != "text" || got.Text.Content != "hello" {
			t.Errorf("body = %+v", got)
		}
	})

	t.Run("api error status", func(t *testing.T) {
		if err := s.SendRobotMsg("bad", "hello"); err == nil {
			t.Error("SendRobotMsg with 400 response should return error")
		}
	})
}

// assertError is a test error type
type assertError string

func (e assertError) Error() string {
	return string(e)
}
