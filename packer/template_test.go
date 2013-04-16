package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func TestParseTemplate_Basic(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": []
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")
	assert.NotNil(result, "template should not be nil")
	assert.Equal(result.Name, "my-image", "name should be correct")
	assert.Length(result.Builders, 0, "no builders")
}

func TestParseTemplate_Invalid(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Note there is an extra comma below for a purposeful
	// syntax error in the JSON.
	data := `
	{
		"name": "my-image",,
		"builders": []
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have an error")
	assert.Nil(result, "should have no result")
}

func TestParseTemplate_BuilderWithoutName(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"type": "amazon-ebs"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")
	assert.NotNil(result, "template should not be nil")
	assert.Length(result.Builders, 1, "should have one builder")

	builder, ok := result.Builders["amazon-ebs"]
	assert.True(ok, "should have amazon-ebs builder")
	assert.Equal(builder.builderName, "amazon-ebs", "builder should be amazon-ebs")
}

func TestParseTemplate_BuilderWithName(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "bob",
				"type": "amazon-ebs"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")
	assert.NotNil(result, "template should not be nil")
	assert.Length(result.Builders, 1, "should have one builder")

	builder, ok := result.Builders["bob"]
	assert.True(ok, "should have bob builder")
	assert.Equal(builder.builderName, "amazon-ebs", "builder should be amazon-ebs")
}
