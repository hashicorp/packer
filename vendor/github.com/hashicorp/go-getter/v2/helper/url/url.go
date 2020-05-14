package url

import (
	"net/url"
)

// Parse parses rawURL into a URL structure.
// The rawURL may be relative or absolute.
//
// Parse is a wrapper for the Go stdlib net/url Parse function, but returns
// Windows "safe" URLs on Windows platforms.
func Parse(rawURL string) (*url.URL, error) {
	return parse(rawURL)
}

// MustParse calls Parse but panics if it errors; it should be mainly used to
// make tests shorter.
func MustParse(s string) *url.URL {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}
