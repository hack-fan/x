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
