package common

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ScrubConfig is a helper that returns a string representation of
// any struct with the given values stripped out.
func ScrubConfig(target interface{}, values ...string) string {
	conf := fmt.Sprintf("Config: %+v", target)
	for _, value := range values {
		if value == "" {
			continue
		}
		conf = strings.Replace(conf, value, "<Filtered>", -1)
	}
	return conf
}

// ChooseString returns the first non-empty value.
func ChooseString(vals ...string) string {
	for _, el := range vals {
		if el != "" {
			return el
		}
	}

	return ""
}

// DownloadableURL processes a URL that may also be a file path and returns
// a completely valid URL. For example, the original URL might be "local/file.iso"
// which isn't a valid URL. DownloadableURL will return "file:///local/file.iso"
func DownloadableURL(original string) (string, error) {
	if runtime.GOOS == "windows" {
		// If the distance to the first ":" is just one character, assume
		// we're dealing with a drive letter and thus a file path.
		idx := strings.Index(original, ":")
		if idx == 1 {
			original = "file:///" + original
		}
	}

	url, err := url.Parse(original)
	if err != nil {
		return "", err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	if url.Scheme == "file" {
		// Windows file handling is all sorts of tricky...
		if runtime.GOOS == "windows" {
			// If the path is using Windows-style slashes, URL parses
			// it into the host field.
			if url.Path == "" && strings.Contains(url.Host, `\`) {
				url.Path = url.Host
				url.Host = ""
			}

			// For Windows absolute file paths, remove leading / prior to processing
			// since net/url turns "C:/" into "/C:/"
			if len(url.Path) > 0 && url.Path[0] == '/' {
				url.Path = url.Path[1:len(url.Path)]
			}
		}

		// Only do the filepath transformations if the file appears
		// to actually exist.
		if _, err := os.Stat(url.Path); err == nil {
			url.Path, err = filepath.Abs(url.Path)
			if err != nil {
				return "", err
			}

			url.Path, err = filepath.EvalSymlinks(url.Path)
			if err != nil {
				return "", err
			}

			url.Path = filepath.Clean(url.Path)
		}

		if runtime.GOOS == "windows" {
			// Also replace all backslashes with forwardslashes since Windows
			// users are likely to do this but the URL should actually only
			// contain forward slashes.
			url.Path = strings.Replace(url.Path, `\`, `/`, -1)
		}
	}

	// Make sure it is lowercased
	url.Scheme = strings.ToLower(url.Scheme)

	// This is to work around issue #5927. This can safely be removed once
	// we distribute with a version of Go that fixes that bug.
	//
	// See: https://code.google.com/p/go/issues/detail?id=5927
	if url.Path != "" && url.Path[0] != '/' {
		url.Path = "/" + url.Path
	}

	// Verify that the scheme is something we support in our common downloader.
	supported := []string{"file", "http", "https"}
	found := false
	for _, s := range supported {
		if url.Scheme == s {
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("Unsupported URL scheme: %s", url.Scheme)
	}

	return url.String(), nil
}
