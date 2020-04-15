package common

import "strings"

// StringsContains returns true if the `haystack` contains the `needle`. Search is case insensitive.
func StringsContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}
