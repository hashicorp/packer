package common

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"sort"
	"strings"
)

// CheckUnusedConfig is a helper that makes sure that the there are no
// unused configuration keys, properly ignoring keys that don't matter.
func CheckUnusedConfig(md *mapstructure.Metadata) error {
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
		Metadata: &md,
		Result:   target,
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
