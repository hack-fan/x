package xtype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Strings is string array for connecting mysql types and json
type Strings []string

func (t Strings) String() string {
	if t == nil {
		return ""
	}
	tmp, _ := json.Marshal([]string(t))
	return string(tmp)
}

func (t Strings) Contains(s string) bool {
	for _, item := range t {
		if item == s {
			return true
		}
	}
	return false
}

func (t Strings) Intersectant(s Strings) bool {
	for _, ss := range s {
		if t.Contains(ss) {
			return true
		}
	}
	return false
}

// SAdd add as set
func (t Strings) SAdd(s string) Strings {
	if t.Contains(s) {
		return t
	}
	return append(t, s)
}

// Remove item which value is s
func (t Strings) Remove(s string) Strings {
	tmp := make(Strings, 0, len(t))
	for _, tt := range t {
		if tt != s {
			tmp = append(tmp, tt)
		}
	}
	return tmp
}

func (t Strings) Union(s Strings) Strings {
	for _, tt := range t {
		s = s.SAdd(tt)
	}
	return s
}

func (t Strings) Sub(s Strings) Strings {
	tmp := t
	for _, ss := range s {
		tmp = tmp.Remove(ss)
	}
	return tmp
}

// MarshalJSON interface
func (t Strings) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(t))
}

// UnmarshalJSON interface
func (t *Strings) UnmarshalJSON(data []byte) error {
	var tmp []string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*t = tmp
	return nil
}

// Scan implements the Scanner interface.
func (t *Strings) Scan(src interface{}) error {
	*t = make([]string, 0)
	if src == nil {
		return nil
	}
	tmp, ok := src.([]byte)
	if !ok {
		return errors.New("read json string array from DB failed")
	}
	if len(tmp) == 0 {
		return nil
	}
	return t.UnmarshalJSON(tmp)
}

// Value implements the driver Valuer interface.
func (t Strings) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return t.String(), nil
}

// Numbers is int array for connecting mysql types and json
type Numbers []int

func (t Numbers) String() string {
	if t == nil {
		return ""
	}
	tmp, _ := json.Marshal([]int(t))
	return string(tmp)
}

func (t Numbers) Contains(s int) bool {
	for _, item := range t {
		if item == s {
			return true
		}
	}
	return false
}

func (t Numbers) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]int(t))
}

func (t *Numbers) UnmarshalJSON(data []byte) error {
	var tmp []int
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*t = tmp
	return nil
}

// Scan implements the Scanner interface.
func (t *Numbers) Scan(src interface{}) error {
	*t = make([]int, 0)
	if src == nil {
		return nil
	}
	tmp, ok := src.([]byte)
	if !ok {
		return errors.New("read json int array from DB failed")
	}
	if len(tmp) == 0 {
		return nil
	}
	return t.UnmarshalJSON(tmp)
}

// Value implements the driver Valuer interface.
func (t Numbers) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return t.String(), nil
}
