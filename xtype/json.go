package xtype

import (
	"encoding/json"
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
