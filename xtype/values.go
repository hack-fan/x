package xtype

import "cmp"

// Compare returns -1, 0, or 1 depending on ordering of a and b.
//
// Deprecated: use [cmp.Compare] directly.
//
//go:fix inline
func Compare[T cmp.Ordered](a, b T) int {
	return cmp.Compare(a, b)
}

// Min returns the minimum of two ordered values.
//
// Deprecated: use the built-in min function.
//
//go:fix inline
func Min[T cmp.Ordered](a, b T) T {
	return min(a, b)
}

// Max returns the maximum of two ordered values.
//
// Deprecated: use the built-in max function.
//
//go:fix inline
func Max[T cmp.Ordered](a, b T) T {
	return max(a, b)
}

// Clamp returns v clamped between lo and hi.
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Coalesce returns the first non-zero value.
//
// Deprecated: use [cmp.Or] directly.
//
//go:fix inline
func Coalesce[T comparable](vals ...T) T {
	return cmp.Or(vals...)
}

// Ptr returns a pointer to v.
//
// Deprecated: use new(v), supported since Go 1.26.
//
//go:fix inline
func Ptr[T any](v T) *T {
	return new(v)
}

// Deref returns the value pointed to, or the zero value if p is nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value pointed to, or def if p is nil.
func DerefOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
