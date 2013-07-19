package common

import (
	"github.com/mitchellh/mapstructure"
)

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
