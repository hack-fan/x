package xtype

import (
	"testing"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"a less than b", 1, 2, -1},
		{"a equal to b", 5, 5, 0},
		{"a greater than b", 10, 5, 1},
		{"negative numbers", -5, -3, -1},
		{"zero", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Compare(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// Test with different types
	t.Run("strings", func(t *testing.T) {
		if Compare("a", "b") != -1 {
			t.Errorf("Compare(\"a\", \"b\") should be -1")
		}
		if Compare("b", "a") != 1 {
			t.Errorf("Compare(\"b\", \"a\") should be 1")
		}
		if Compare("same", "same") != 0 {
			t.Errorf("Compare(\"same\", \"same\") should be 0")
		}
	})

	t.Run("floats", func(t *testing.T) {
		if Compare(1.5, 2.5) != -1 {
			t.Errorf("Compare(1.5, 2.5) should be -1")
		}
		if Compare(3.14, 3.14) != 0 {
			t.Errorf("Compare(3.14, 3.14) should be 0")
		}
	})
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"first smaller", 1, 5, 1},
		{"second smaller", 10, 2, 2},
		{"equal", 7, 7, 7},
		{"negative", -5, 5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Min(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// Test with different types
	t.Run("strings", func(t *testing.T) {
		if Min("a", "b") != "a" {
			t.Errorf("Min(\"a\", \"b\") should be \"a\"")
		}
	})

	t.Run("floats", func(t *testing.T) {
		if Min(1.1, 2.2) != 1.1 {
			t.Errorf("Min(1.1, 2.2) should be 1.1")
		}
	})
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"first larger", 10, 5, 10},
		{"second larger", 2, 8, 8},
		{"equal", 7, 7, 7},
		{"negative", -1, -5, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Max(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Max(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}

	// Test with different types
	t.Run("strings", func(t *testing.T) {
		if Max("a", "b") != "b" {
			t.Errorf("Max(\"a\", \"b\") should be \"b\"")
		}
	})

	t.Run("floats", func(t *testing.T) {
		if Max(1.1, 2.2) != 2.2 {
			t.Errorf("Max(1.1, 2.2) should be 2.2")
		}
	})
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name            string
		value, min, max int
		expected        int
	}{
		{"value in range", 5, 1, 10, 5},
		{"value below min", 0, 1, 10, 1},
		{"value above max", 15, 1, 10, 10},
		{"value equals min", 1, 1, 10, 1},
		{"value equals max", 10, 1, 10, 10},
		{"negative range", -2, -5, -1, -2},
		{"clamped to negative", -10, -5, 5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.value, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d, want %d",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}

	// Test with different types
	t.Run("floats", func(t *testing.T) {
		if Clamp(5.5, 1.0, 10.0) != 5.5 {
			t.Errorf("Clamp(5.5, 1.0, 10.0) should be 5.5")
		}
		if Clamp(0.5, 1.0, 10.0) != 1.0 {
			t.Errorf("Clamp(0.5, 1.0, 10.0) should be 1.0")
		}
	})
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		vals     []int
		expected int
	}{
		{"first non-zero", []int{5, 0, 10}, 5},
		{"second non-zero", []int{0, 7, 0}, 7},
		{"last non-zero", []int{0, 0, 3}, 3},
		{"all zero", []int{0, 0, 0}, 0},
		{"single value", []int{42}, 42},
		{"single zero", []int{0}, 0},
		{"empty", []int{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Coalesce(tt.vals...)
			if result != tt.expected {
				t.Errorf("Coalesce(%v) = %d, want %d", tt.vals, result, tt.expected)
			}
		})
	}

	// Test with strings
	t.Run("strings", func(t *testing.T) {
		if Coalesce("", "b", "c") != "b" {
			t.Errorf("Coalesce(\"\", \"b\", \"c\") should be \"b\"")
		}
		if Coalesce("a", "b", "c") != "a" {
			t.Errorf("Coalesce(\"a\", \"b\", \"c\") should be \"a\"")
		}
		if Coalesce("", "", "") != "" {
			t.Errorf("Coalesce(\"\", \"\", \"\") should be \"\"")
		}
	})

	// Test with pointers
	t.Run("pointers", func(t *testing.T) {
		a := ptrInt(42)
		b := ptrInt(100)
		if Coalesce[*int](nil, a, b) != a {
			t.Errorf("Coalesce(nil, a, b) should return a")
		}
		if Coalesce[*int](nil, nil, b) != b {
			t.Errorf("Coalesce(nil, nil, b) should return b")
		}
	})
}

func TestPtr(t *testing.T) {
	t.Run("int pointer", func(t *testing.T) {
		p := Ptr(42)
		if p == nil {
			t.Errorf("Ptr(42) returned nil")
		}
		if *p != 42 {
			t.Errorf("Ptr(42) = %d, want 42", *p)
		}
	})

	t.Run("string pointer", func(t *testing.T) {
		p := Ptr("hello")
		if p == nil {
			t.Errorf("Ptr(\"hello\") returned nil")
		}
		if *p != "hello" {
			t.Errorf("Ptr(\"hello\") = %s, want \"hello\"", *p)
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type S struct{ Field int }
		p := Ptr(S{Field: 123})
		if p == nil {
			t.Errorf("Ptr(S{Field: 123}) returned nil")
		}
		if p.Field != 123 {
			t.Errorf("Ptr(S{Field: 123}).Field = %d, want 123", p.Field)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		p := Ptr(0)
		if p == nil {
			t.Errorf("Ptr(0) returned nil")
		}
		if *p != 0 {
			t.Errorf("Ptr(0) = %d, want 0", *p)
		}
	})
}

func TestDeref(t *testing.T) {
	t.Run("int pointer", func(t *testing.T) {
		val := 42
		result := Deref(&val)
		if result != 42 {
			t.Errorf("Deref(&42) = %d, want 42", result)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var p *int
		result := Deref(p)
		if result != 0 {
			t.Errorf("Deref(nil) = %d, want 0", result)
		}
	})

	t.Run("string pointer", func(t *testing.T) {
		s := "hello"
		result := Deref(&s)
		if result != "hello" {
			t.Errorf("Deref(&\"hello\") = %s, want \"hello\"", result)
		}
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var p *string
		result := Deref(p)
		if result != "" {
			t.Errorf("Deref(nil *string) = %s, want \"\"", result)
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type S struct{ Field int }
		s := S{Field: 123}
		result := Deref(&s)
		if result.Field != 123 {
			t.Errorf("Deref(&S{Field: 123}).Field = %d, want 123", result.Field)
		}
	})

	t.Run("nil struct pointer", func(t *testing.T) {
		type S struct{ Field int }
		var p *S
		result := Deref(p)
		if result.Field != 0 {
			t.Errorf("Deref(nil *S).Field = %d, want 0", result.Field)
		}
	})
}

func TestDerefOr(t *testing.T) {
	t.Run("int pointer with value", func(t *testing.T) {
		val := 42
		result := DerefOr(&val, 99)
		if result != 42 {
			t.Errorf("DerefOr(&42, 99) = %d, want 42", result)
		}
	})

	t.Run("nil int pointer", func(t *testing.T) {
		var p *int
		result := DerefOr(p, 99)
		if result != 99 {
			t.Errorf("DerefOr(nil, 99) = %d, want 99", result)
		}
	})

	t.Run("string pointer with value", func(t *testing.T) {
		s := "hello"
		result := DerefOr(&s, "fallback")
		if result != "hello" {
			t.Errorf("DerefOr(&\"hello\", \"fallback\") = %s, want \"hello\"", result)
		}
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var p *string
		result := DerefOr(p, "fallback")
		if result != "fallback" {
			t.Errorf("DerefOr(nil, \"fallback\") = %s, want \"fallback\"", result)
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type S struct{ Field int }
		s := S{Field: 123}
		def := S{Field: 999}
		result := DerefOr(&s, def)
		if result.Field != 123 {
			t.Errorf("DerefOr(&S{123}, S{999}).Field = %d, want 123", result.Field)
		}
	})

	t.Run("nil struct pointer", func(t *testing.T) {
		type S struct{ Field int }
		def := S{Field: 999}
		result := DerefOr((*S)(nil), def)
		if result.Field != 999 {
			t.Errorf("DerefOr(nil, S{999}).Field = %d, want 999", result.Field)
		}
	})

	t.Run("zero default", func(t *testing.T) {
		var p *int
		result := DerefOr(p, 0)
		if result != 0 {
			t.Errorf("DerefOr(nil, 0) = %d, want 0", result)
		}
	})
}

func ptrInt(v int) *int {
	return &v
}
