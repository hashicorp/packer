package main

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func TestConfig_MergeConfig(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	aString := `
	[commands]
	a = "1"
	b = "1"
	`

	bString := `
	[commands]
	a = "1"
	b = "2"
	c = "3"
	`

	a, _ := parseConfig(aString)
	b, _ := parseConfig(bString)
	result := mergeConfig(a, b)

	assert.Equal(result.Commands["a"], "1", "a should be 1")
	assert.Equal(result.Commands["b"], "2", "a should be 2")
	assert.Equal(result.Commands["c"], "3", "a should be 3")
}

func TestConfig_ParseConfig_Bad(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	[commands]
	foo = bar
	`

	_, err := parseConfig(data)
	assert.NotNil(err, "should have an error")
}

func TestConfig_ParseConfig_DefaultConfig(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := parseConfig(defaultConfig)
	assert.Nil(err, "should be able to parse the default config")
}

func TestConfig_ParseConfig_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	[commands]
	foo = "bar"
	`

	c, err := parseConfig(data)
	assert.Nil(err, "should not have an error")
	assert.Equal(c.CommandNames(), []string{"foo"}, "should have correct command names")
	assert.Equal(c.Commands["foo"], "bar", "should have the command")
}
