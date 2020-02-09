package rpc

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
)

// cty.Value is does not know how to encode itself through the wire so we
// transform it to bytes.
func encodeCTYValues(config []interface{}) ([]interface{}, error) {
	for i := range config {
		if v, ok := config[i].(cty.Value); ok {
			b, err := json.Marshal(v, cty.DynamicPseudoType)
			if err != nil {
				return nil, err
			}
			config[i] = b
		}
	}
	return config, nil
}

// decodeCTYValues will try to decode a cty value when it finds a byte slice
func decodeCTYValues(config []interface{}) ([]interface{}, error) {
	for i := range config {
		if b, ok := config[i].([]byte); ok {
			t, err := json.Unmarshal(b, cty.DynamicPseudoType)
			if err != nil {
				return nil, err
			}
			config[i] = t
		}
	}
	return config, nil
}
