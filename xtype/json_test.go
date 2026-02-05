package xtype

import (
	"testing"
)

func TestToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "number",
			input:    42,
			expected: `42`,
		},
		{
			name:     "float",
			input:    3.14,
			expected: `3.14`,
		},
		{
			name:     "bool",
			input:    true,
			expected: `true`,
		},
		{
			name:     "slice",
			input:    []int{1, 2, 3},
			expected: `[1,2,3]`,
		},
		{
			name:     "map",
			input:    map[string]int{"a": 1, "b": 2},
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "struct",
			input:    struct{ Name string }{Name: "test"},
			expected: `{"Name":"test"}`,
		},
		{
			name:     "nil",
			input:    nil,
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSON(tt.input)
			if result != tt.expected {
				t.Errorf("ToJSON(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToJSON_invalid(t *testing.T) {
	// Function with unmarshalable type should return empty string
	result := ToJSON(func() {})
	if result != "" {
		t.Errorf("ToJSON(func) should return empty string, got %q", result)
	}
}

func TestToJSONPretty(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		contains []string // check if output contains these strings
	}{
		{
			name:     "simple object",
			input:    map[string]int{"a": 1, "b": 2},
			contains: []string{"{", "  ", "a", "b", "}"},
		},
		{
			name:     "nested object",
			input:    map[string]any{"outer": map[string]int{"inner": 1}},
			contains: []string{"outer", "inner"},
		},
		{
			name:     "array",
			input:    []int{1, 2, 3},
			contains: []string{"[", "]", "1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToJSONPretty(tt.input)
			if result == "" {
				t.Errorf("ToJSONPretty(%v) returned empty string", tt.input)
				return
			}
			for _, substr := range tt.contains {
				if !contains(result, substr) {
					t.Errorf("ToJSONPretty output should contain %q, got %q", substr, result)
				}
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		data    string
		target  any
		wantErr bool
		check   func(any) bool
	}{
		{
			name:    "valid object",
			data:    `{"name":"Alice","age":30}`,
			target:  &Person{},
			wantErr: false,
			check: func(p any) bool {
				person := p.(*Person)
				return person.Name == "Alice" && person.Age == 30
			},
		},
		{
			name:    "valid array",
			data:    `[1,2,3]`,
			target:  &[]int{},
			wantErr: false,
			check: func(p any) bool {
				arr := p.(*[]int)
				return len(*arr) == 3 && (*arr)[0] == 1 && (*arr)[1] == 2 && (*arr)[2] == 3
			},
		},
		{
			name:    "valid string",
			data:    `"hello"`,
			target:  new(string),
			wantErr: false,
			check: func(p any) bool {
				s := p.(*string)
				return *s == "hello"
			},
		},
		{
			name:    "valid number",
			data:    `42`,
			target:  new(int),
			wantErr: false,
			check: func(p any) bool {
				i := p.(*int)
				return *i == 42
			},
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			target:  &Person{},
			wantErr: true,
		},
		{
			name:    "empty string",
			data:    "",
			target:  &Person{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FromJSON(tt.data, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(tt.target) {
				t.Errorf("FromJSON() check failed for %v", tt.target)
			}
		})
	}
}

func TestFromJSONMust(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name   string
		data   string
		target any
		check  func(any) bool
	}{
		{
			name:   "valid JSON",
			data:   `{"name":"Bob"}`,
			target: &Person{},
			check: func(p any) bool {
				person := p.(*Person)
				return person.Name == "Bob"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("FromJSONMust() panicked unexpectedly: %v", r)
				}
			}()
			FromJSONMust(tt.data, tt.target)
			if tt.check != nil && !tt.check(tt.target) {
				t.Errorf("FromJSONMust() check failed for %v", tt.target)
			}
		})
	}

	t.Run("invalid JSON panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("FromJSONMust() should panic on invalid JSON")
			}
		}()
		var p Person
		FromJSONMust("{invalid}", &p)
	})
}

func TestToMap(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name    string
		input   any
		wantErr bool
		check   func(map[string]any) bool
	}{
		{
			name:    "struct to map",
			input:   Person{Name: "Alice", Age: 30},
			wantErr: false,
			check: func(m map[string]any) bool {
				return m["name"] == "Alice" && m["age"] == float64(30)
			},
		},
		{
			name:    "map to map",
			input:   map[string]any{"key": "value"},
			wantErr: false,
			check: func(m map[string]any) bool {
				return m["key"] == "value"
			},
		},
		{
			name:    "nested struct",
			input:   struct{ Inner Person }{Inner: Person{Name: "Bob", Age: 25}},
			wantErr: false,
			check: func(m map[string]any) bool {
				inner, ok := m["Inner"].(map[string]any)
				if !ok {
					return false
				}
				return inner["name"] == "Bob" && inner["age"] == float64(25)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Errorf("ToMap() check failed for %v", result)
			}
		})
	}
}

func TestToMap_invalid(t *testing.T) {
	// Channel cannot be marshaled to JSON
	result, err := ToMap(make(chan int))
	if err == nil {
		t.Errorf("ToMap(channel) should return error, got nil")
	}
	if result != nil {
		t.Errorf("ToMap(channel) result should be nil, got %v", result)
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
