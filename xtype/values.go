package xtype

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Values maps a string key to a list of values.
// It is typically used for object key.
type Values map[string]string

// Get gets the value of the given key.
// If there are no values associated with the key, Get returns
// the empty string.
func (v Values) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs
}

// Set sets the key to value. It replaces any existing values.
// don't contain _ and - in values, use [a-z0-9] only
func (v Values) Set(key, value string) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Values) Del(key string) {
	delete(v, key)
}

// Parse parses the encoded values string
func ParseQuery(src string) (Values, error) {
	err := fmt.Errorf("invalid value string: %s", src)
	m := make(Values)
	pairs := strings.Split(src, "_")
	for _, str := range pairs {
		pair := strings.Split(str, "-")
		if len(pair) != 2 {
			return nil, err
		}
		k, err := url.PathUnescape(pair[0])
		if err != nil {
			return nil, err
		}
		v, err := url.PathUnescape(pair[1])
		if err != nil {
			return nil, err
		}
		m.Set(k, v)
	}
	return m, nil
}

// Encode encodes the values into URL encoded form
// ("bar-baz_foo-quux") sorted by key.
func (v Values) Encode() string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		if buf.Len() > 0 {
			buf.WriteByte('_')
		}
		buf.WriteString(url.PathEscape(k))
		buf.WriteByte('-')
		buf.WriteString(url.PathEscape(vs))
	}
	return buf.String()
}
