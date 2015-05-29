package atlas

import (
	"math"
	"path/filepath"
	"strings"
)

// longestCommonPrefix finds the longest common prefix for all the strings
// given as an argument, or returns the empty string if a prefix can't be
// found.
//
// This function just uses brute force instead of a more optimized algorithm.
func longestCommonPrefix(vs []string) string {
	var length int64
	// Find the shortest string
	var shortest string
	length = math.MaxUint32
	for _, v := range vs {
		if int64(len(v)) < length {
			shortest = v
			length = int64(len(v))
		}
	}

	// Now go through and find a prefix to all the strings using this
	// short string, which itself must contain the prefix.
	for i := len(shortest); i > 0; i-- {
		// We only care about prefixes with path seps
		if shortest[i-1] != filepath.Separator {
			continue
		}

		bad := false
		prefix := shortest[0:i]
		for _, v := range vs {
			if !strings.HasPrefix(v, prefix) {
				bad = true
				break
			}
		}

		if !bad {
			return prefix
		}
	}

	return ""
}
