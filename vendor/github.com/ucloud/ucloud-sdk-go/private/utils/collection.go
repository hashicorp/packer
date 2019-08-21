package utils

import (
	"fmt"
	"strings"
)

// MergeMap will merge two map and return a new map
func MergeMap(args ...map[string]string) map[string]string {
	m := map[string]string{}
	for _, kv := range args {
		for k, v := range kv {
			m[k] = v
		}
	}
	return m
}

// SetMapIfNotExists will set a key-value of the map if the key is not exists
func SetMapIfNotExists(m map[string]string, k string, v string) {
	if _, ok := m[k]; !ok && v != "" {
		m[k] = v
	}
}

// IsStringIn will return if the value is contains by an array
func IsStringIn(val string, availables []string) bool {
	for _, choice := range availables {
		if val == choice {
			return true
		}
	}

	return false
}

// CheckStringIn will check if the value is contains by an array
func CheckStringIn(val string, availables []string) error {
	if IsStringIn(val, availables) {
		return nil
	}
	return fmt.Errorf("got %s, should be one of %s", val, strings.Join(availables, ","))
}
