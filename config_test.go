package main

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func TestConfig_ParseConfig_Bad(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	[commands]
	foo = bar
	`

	_, err := parseConfig(data)
	assert.NotNil(err, "should have an error")
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
