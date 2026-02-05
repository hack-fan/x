package xerr

import (
	"errors"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(404, "NotFound", "User not found")
	if err.StatusCode() != 404 {
		t.Errorf("Expected status code 404, got %d", err.StatusCode())
	}
	if err.Key != "NotFound" {
		t.Errorf("Expected key NotFound, got %s", err.Key)
	}
	if err.Message != "User not found" {
		t.Errorf("Expected message 'User not found', got %s", err.Message)
	}
}

func TestNewf(t *testing.T) {
	err := Newf(500, "DatabaseError", "failed to connect: %s", "timeout")
	if err.StatusCode() != 500 {
		t.Errorf("Expected status code 500, got %d", err.StatusCode())
	}
	if err.Key != "DatabaseError" {
		t.Errorf("Expected key DatabaseError, got %s", err.Key)
	}
	expectedMsg := "failed to connect: timeout"
	if err.Message != expectedMsg {
		t.Errorf("Expected message '%s', got %s", expectedMsg, err.Message)
	}
}

func TestNewfWithAny(t *testing.T) {
	// Test that Newf accepts any type with the modern 'any' type
	err := Newf(400, "InvalidInput", "invalid value: %v", 12345)
	if err.StatusCode() != 400 {
		t.Errorf("Expected status code 400, got %d", err.StatusCode())
	}
}

func TestError_Error(t *testing.T) {
	err := New(404, "NotFound", "User not found")
	if err.Error() != "User not found" {
		t.Errorf("Expected Error() to return 'User not found', got %s", err.Error())
	}
}

func TestError_Unwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := &Error{
		err:     baseErr,
		code:    500,
		Key:     "Wrapped",
		Message: "wrapped error",
	}

	if err.Unwrap() != baseErr {
		t.Error("Unwrap() should return the base error")
	}
}

func TestError_UnwrapAll(t *testing.T) {
	t.Run("no wrapped errors", func(t *testing.T) {
		err := New(404, "NotFound", "not found")
		unwrapped := err.UnwrapAll()
		if len(unwrapped) != 0 {
			t.Errorf("Expected 0 unwrapped errors, got %d", len(unwrapped))
		}
	})

	t.Run("single wrapped error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := &Error{
			err:     baseErr,
			code:    500,
			Key:     "Wrapped",
			Message: "wrapped error",
		}
		unwrapped := err.UnwrapAll()
		if len(unwrapped) != 1 {
			t.Errorf("Expected 1 unwrapped error, got %d", len(unwrapped))
		}
		if unwrapped[0] != baseErr {
			t.Error("UnwrapAll() should return the base error")
		}
	})

	t.Run("multiple wrapped errors", func(t *testing.T) {
		baseErr := errors.New("base error")
		middleErr := fmt.Errorf("middle: %w", baseErr)
		topErr := &Error{
			err:     middleErr,
			code:    500,
			Key:     "Top",
			Message: "top error",
		}

		unwrapped := topErr.UnwrapAll()
		if len(unwrapped) != 2 {
			t.Errorf("Expected 2 unwrapped errors, got %d", len(unwrapped))
		}
		if unwrapped[0] != middleErr {
			t.Error("First unwrapped error should be middleErr")
		}
		if unwrapped[1] != baseErr {
			t.Error("Second unwrapped error should be baseErr")
		}
	})
}

func TestIs(t *testing.T) {
	err := New(404, "NotFound", "User not found")

	if !Is(err, "NotFound") {
		t.Error("Is() should return true for matching key")
	}

	if Is(err, "OtherError") {
		t.Error("Is() should return false for non-matching key")
	}

	if Is(errors.New("standard error"), "NotFound") {
		t.Error("Is() should return false for non-xerr errors")
	}
}

func TestIsCode(t *testing.T) {
	err := New(404, "NotFound", "User not found")

	if !IsCode(err, 404) {
		t.Error("IsCode() should return true for matching code")
	}

	if IsCode(err, 500) {
		t.Error("IsCode() should return false for non-matching code")
	}

	if IsCode(errors.New("standard error"), 404) {
		t.Error("IsCode() should return false for non-xerr errors")
	}
}

func TestAs(t *testing.T) {
	xerr := New(404, "NotFound", "User not found")

	result, ok := As(xerr)
	if !ok {
		t.Error("As() should return true for xerr errors")
	}
	if result == nil {
		t.Fatal("As() should return the error")
	}
	if result.Key != "NotFound" {
		t.Errorf("Expected key NotFound, got %s", result.Key)
	}

	// Test with non-xerr error
	result, ok = As(errors.New("standard error"))
	if ok {
		t.Error("As() should return false for non-xerr errors")
	}
	if result != nil {
		t.Error("As() should return nil for non-xerr errors")
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "400 Bad Request",
			err:      New(400, "BadRequest", "bad request"),
			expected: true,
		},
		{
			name:     "404 Not Found",
			err:      New(404, "NotFound", "not found"),
			expected: true,
		},
		{
			name:     "499 Client Closed Request",
			err:      New(499, "ClientClosed", "client closed"),
			expected: true,
		},
		{
			name:     "500 Internal Server Error",
			err:      New(500, "ServerError", "server error"),
			expected: false,
		},
		{
			name:     "200 OK",
			err:      New(200, "OK", "success"),
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsClientError(tt.err)
			if result != tt.expected {
				t.Errorf("IsClientError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "500 Internal Server Error",
			err:      New(500, "InternalServerError", "internal error"),
			expected: true,
		},
		{
			name:     "502 Bad Gateway",
			err:      New(502, "BadGateway", "bad gateway"),
			expected: true,
		},
		{
			name:     "599 Network Connect Timeout",
			err:      New(599, "NetworkTimeout", "timeout"),
			expected: true,
		},
		{
			name:     "404 Not Found",
			err:      New(404, "NotFound", "not found"),
			expected: false,
		},
		{
			name:     "400 Bad Request",
			err:      New(400, "BadRequest", "bad request"),
			expected: false,
		},
		{
			name:     "200 OK",
			err:      New(200, "OK", "success"),
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsServerError(tt.err)
			if result != tt.expected {
				t.Errorf("IsServerError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	t.Run("join multiple errors", func(t *testing.T) {
		err1 := errors.New("error 1")
		err2 := errors.New("error 2")
		err3 := errors.New("error 3")

		joined := Join(500, "MultipleErrors", "multiple errors occurred", err1, err2, err3)

		if joined.StatusCode() != 500 {
			t.Errorf("Expected status code 500, got %d", joined.StatusCode())
		}
		if joined.Key != "MultipleErrors" {
			t.Errorf("Expected key MultipleErrors, got %s", joined.Key)
		}
		if joined.Message != "multiple errors occurred" {
			t.Errorf("Expected message 'multiple errors occurred', got %s", joined.Message)
		}

		// Verify that errors.Is can find the joined errors
		if !errors.Is(joined, err1) {
			t.Error("errors.Is should find err1 in joined error")
		}
		if !errors.Is(joined, err2) {
			t.Error("errors.Is should find err2 in joined error")
		}
		if !errors.Is(joined, err3) {
			t.Error("errors.Is should find err3 in joined error")
		}
	})

	t.Run("join no errors", func(t *testing.T) {
		joined := Join(500, "NoErrors", "no errors provided")

		if joined.StatusCode() != 500 {
			t.Errorf("Expected status code 500, got %d", joined.StatusCode())
		}
	})

	t.Run("join nil errors", func(t *testing.T) {
		joined := Join(500, "NilErrors", "with nil errors", nil, nil)

		if joined.StatusCode() != 500 {
			t.Errorf("Expected status code 500, got %d", joined.StatusCode())
		}
	})

	t.Run("join xerr errors", func(t *testing.T) {
		err1 := New(400, "BadRequest", "invalid input")
		err2 := New(404, "NotFound", "resource not found")

		joined := Join(500, "MultipleErrors", "multiple validation errors", err1, err2)

		if joined.StatusCode() != 500 {
			t.Errorf("Expected status code 500, got %d", joined.StatusCode())
		}

		// Verify that errors.Is can find the wrapped xerr errors
		if !errors.Is(joined, err1) {
			t.Error("errors.Is should find err1 in joined error")
		}
		if !errors.Is(joined, err2) {
			t.Error("errors.Is should find err2 in joined error")
		}
	})
}

func TestParseResp(t *testing.T) {
	// Note: ParseResp requires http.Response which is harder to test
	// This is a minimal test to verify the function exists and handles nil
	t.Run("nil response", func(t *testing.T) {
		err := ParseResp(nil)
		if err == nil {
			t.Error("ParseResp(nil) should return an error")
		}
		if err.StatusCode() != 500 {
			t.Errorf("Expected status code 500 for nil response, got %d", err.StatusCode())
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := Newf(500, "WrappedError", "wrapped: %w", baseErr)

	// Test errors.Is
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find baseErr in wrappedErr")
	}

	// Test As
	var xerr *Error
	if !errors.As(wrappedErr, &xerr) {
		t.Error("errors.As should extract *Error from wrappedErr")
	}
	if xerr.Key != "WrappedError" {
		t.Errorf("Expected key WrappedError, got %s", xerr.Key)
	}
}

func TestServerErrorConstant(t *testing.T) {
	if ServerError.StatusCode() != 500 {
		t.Errorf("Expected ServerError status code 500, got %d", ServerError.StatusCode())
	}
	if ServerError.Key != "ServerError" {
		t.Errorf("Expected ServerError key 'ServerError', got %s", ServerError.Key)
	}
}
