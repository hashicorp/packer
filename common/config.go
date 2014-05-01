package common

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
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

// ChooseString returns the first non-empty value.
func ChooseString(vals ...string) string {
	for _, el := range vals {
		if el != "" {
			return el
		}
	}

	return ""
}

// DecodeConfig is a helper that handles decoding raw configuration using
// mapstructure. It returns the metadata and any errors that may happen.
// If you need extra configuration for mapstructure, you should configure
// it manually and not use this helper function.
func DecodeConfig(target interface{}, raws ...interface{}) (*mapstructure.Metadata, error) {
	decodeHook, err := decodeConfigHook(raws)
	if err != nil {
		return nil, err
	}

	var md mapstructure.Metadata
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			decodeHook,
			mapstructure.StringToSliceHookFunc(","),
		),
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

// This returns a mapstructure.DecodeHookFunc that automatically template
// processes any configuration values that aren't strings but have been
// provided as strings.
//
// For example: "image_id" wants an int and the user uses a string with
// a user variable like "{{user `image_id`}}". This decode hook makes that
// work.
func decodeConfigHook(raws []interface{}) (mapstructure.DecodeHookFunc, error) {
	// First thing we do is decode PackerConfig so that we can have access
	// to the user variables so that we can process some templates.
	var pc PackerConfig

	decoderConfig := &mapstructure.DecoderConfig{
		Result:           &pc,
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}
	for _, raw := range raws {
		if err := decoder.Decode(raw); err != nil {
			return nil, err
		}
	}

	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	tpl.UserVars = pc.PackerUserVars

	return func(f reflect.Kind, t reflect.Kind, v interface{}) (interface{}, error) {
		if t != reflect.String {
			// We need to convert []uint8 to string. We have to do this
			// because internally Packer uses MsgPack for RPC and the MsgPack
			// codec turns strings into []uint8
			if f == reflect.Slice {
				dataVal := reflect.ValueOf(v)
				dataType := dataVal.Type()
				elemKind := dataType.Elem().Kind()
				if elemKind == reflect.Uint8 {
					v = string(dataVal.Interface().([]uint8))
				}
			}

			if sv, ok := v.(string); ok {
				var err error
				v, err = tpl.Process(sv, nil)
				if err != nil {
					return nil, err
				}
			}
		}

		return v, nil
	}, nil
}
