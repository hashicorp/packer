package packer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenConfigKeys_nil(t *testing.T) {
	f := flattenConfigKeys(nil)
	assert.Zero(t, f, "Expected empty list.")
}

func TestFlattenConfigKeys_nested(t *testing.T) {
	inp := make(map[string]interface{})
	inp["A"] = ""
	inp["B"] = []string{}

	c := make(map[string]interface{})
	c["X"] = ""
	d := make(map[string]interface{})
	d["a"] = ""

	c["Y"] = d
	inp["C"] = c

	assert.Equal(t,
		[]string{"A", "B", "C/X", "C/Y/a"},
		flattenConfigKeys(inp),
		"Input didn't flatten correctly.",
	)
}
