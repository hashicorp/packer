package common

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ScrubConfig is a helper that returns a string representation of
// any struct with the given values stripped out.
func ScrubConfig(target interface{}, values ...string) string {
	conf := fmt.Sprintf("Config: %+v", target)
	for _, value := range values {
		conf = strings.Replace(conf, value, "<Filtered>", -1)
	}
	return conf
}

// CheckUnusedConfig is a helper that makes sure that the there are no
// unused configuration keys, properly ignoring keys that don't matter.
func CheckUnusedConfig(md *mapstructure.Metadata) *packer.MultiError {
	errs := make([]error, 0)

	if md.Unused != nil && len(md.Unused) > 0 {
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			if unused != "type" && !strings.HasPrefix(unused, "packer_") {
				errs = append(
					errs, fmt.Errorf("Unknown configuration key: %s", unused))
			}
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

// DecodeConfig is a helper that handles decoding raw configuration using
// mapstructure. It returns the metadata and any errors that may happen.
// If you need extra configuration for mapstructure, you should configure
// it manually and not use this helper function.
func DecodeConfig(target interface{}, raws ...interface{}) (*mapstructure.Metadata, error) {
	var md mapstructure.Metadata
	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         &md,
		Result:           target,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}

	for _, raw := range raws {
		err := decoder.Decode(raw)
		if err != nil {
			return nil, err
		}
	}

	return &md, nil
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
		// For Windows absolute file paths, remove leading / prior to processing
		// since net/url turns "C:/" into "/C:/"
		if runtime.GOOS == "windows" && url.Path[0] == '/' {
			url.Path = url.Path[1:len(url.Path)]

			// Also replace all backslashes with forwardslashes since Windows
			// users are likely to do this but the URL should actually only
			// contain forward slashes.
			url.Path = strings.Replace(url.Path, `\`, `/`, -1)
		}

		if _, err := os.Stat(url.Path); err != nil {
			return "", err
		}

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
