package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func TestFormatName_Basic(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name"
	result, err := FormatName(text)

	assert.Nil(err, "should not error")
	assert.NotEmpty(result, "name should not be empty")
	assert.Equal(result, text, "should be equal")
}

func TestFormatName_BadSub(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name {{ foobar }}"
	result, err := FormatName(text)

	assert.NotNil(err, "should error")
	assert.Empty(result, "name should be empty")
}

func TestFormatName_CreateTime(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name {{ .CreateTime }}"
	result, err := FormatName(text)

	assert.Nil(err, "should not error")
	assert.NotEmpty(result, "name should not be empty")
}

func TestFormatName_Time(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name {{ time \"UTC\" \"2006-01-02T15:04:05Z\"}}"
	result, err := FormatName(text)

	assert.Nil(err, "should not error")
	assert.NotEmpty(result, "name should not be empty")
}

func TestFormatName_BadTime(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name {{ time UTC }}"
	result, err := FormatName(text)

	assert.NotNil(err, "should not error")
	assert.Empty(result, "name should not be empty")
}

func TestFormatName_User(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	text := "packer-name {{ user  }}"
	result, err := FormatName(text)

	assert.Nil(err, "should not error")
	assert.NotEmpty(result, "name should not be empty")
}
