package xecho

import (
	"github.com/labstack/echo/v4"
)

const (
	// ctxStoreKey is the internal key for storing the context store in echo context
	ctxStoreKey = "_xecho_ctx_store"
)

// GetCtx retrieves a value from the echo context with type safety.
// T is the type of the value to retrieve.
// Returns the value and true if found, zero value and false otherwise.
//
// Example:
//
//	userID, ok := GetCtx[string](c, "user_id")
//	if !ok {
//	    return errors.New("user_id not found")
//	}
func GetCtx[T any](c echo.Context, key string) (T, bool) {
	store, ok := c.Get(ctxStoreKey).(map[string]any)
	if !ok {
		var zero T
		return zero, false
	}

	val, ok := store[key].(T)
	if !ok {
		var zero T
		return zero, false
	}

	return val, true
}

// SetCtx stores a value in the echo context with type safety.
// T is the type of the value to store.
//
// Example:
//
//	SetCtx(c, "user_id", "12345")
//	SetCtx(c, "is_admin", true)
func SetCtx[T any](c echo.Context, key string, val T) {
	store, ok := c.Get(ctxStoreKey).(map[string]any)
	if !ok {
		store = make(map[string]any)
		c.Set(ctxStoreKey, store)
	}

	store[key] = val
}

// MustGetCtx retrieves a value from the echo context or panics if not found.
// Use this when you're certain the value exists.
//
// Example:
//
//	userID := MustGetCtx[string](c, "user_id")
func MustGetCtx[T any](c echo.Context, key string) T {
	val, ok := GetCtx[T](c, key)
	if !ok {
		panic("context key not found: " + key)
	}
	return val
}
