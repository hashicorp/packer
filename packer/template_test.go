package packer

import (
	"cgl.tideland.biz/asserts"
	"sort"
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

func TestParseTemplate_BuilderWithoutType(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [{}]
	}
	`

	_, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have error")
}

func TestParseTemplate_BuilderWithNonStringType(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [{
			"type": 42
		}]
	}
	`

	_, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have error")
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
	assert.Equal(builder.builderType, "amazon-ebs", "builder should be amazon-ebs")
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
	assert.Equal(builder.builderType, "amazon-ebs", "builder should be amazon-ebs")
}

func TestParseTemplate_BuilderWithConflictingName(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "bob",
				"type": "amazon-ebs"
			},
			{
				"name": "bob",
				"type": "foo",
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have error")
}

func TestParseTemplate_Hooks(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",

		"hooks": {
			"event": ["foo", "bar"]
		}
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")
	assert.NotNil(result, "template should not be nil")
	assert.Length(result.Hooks, 1, "should have one hook")

	hooks, ok := result.Hooks["event"]
	assert.True(ok, "should have hook")
	assert.Equal(hooks, []string{"foo", "bar"}, "hooks should be correct")
}

func TestParseTemplate_ProvisionerWithoutType(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"provisioners": [{}]
	}
	`

	_, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have error")
}

func TestParseTemplate_ProvisionerWithNonStringType(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"provisioners": [{
			"type": 42
		}]
	}
	`

	_, err := ParseTemplate([]byte(data))
	assert.NotNil(err, "should have error")
}

func TestParseTemplate_Provisioners(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"provisioners": [
			{
				"type": "shell"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")
	assert.NotNil(result, "template should not be nil")
	assert.Length(result.Provisioners, 1, "should have one provisioner")
	assert.Equal(result.Provisioners[0].pType, "shell", "provisioner should be shell")
}

func TestTemplate_BuildNames(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "bob",
				"type": "amazon-ebs"
			},
			{
				"name": "chris",
				"type": "another"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")

	buildNames := result.BuildNames()
	sort.Strings(buildNames)
	assert.Equal(buildNames, []string{"bob", "chris"}, "should have proper builds")
}

func TestTemplate_BuildUnknown(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")

	build, err := template.Build("nope", nil)
	assert.Nil(build, "build should be nil")
	assert.NotNil(err, "should have error")
}

func TestTemplate_BuildUnknownBuilder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")

	builderFactory := func(string) (Builder, error) { return nil, nil }
	components := &ComponentFinder{Builder: builderFactory}
	build, err := template.Build("test1", components)
	assert.Nil(build, "build should be nil")
	assert.NotNil(err, "should have error")
}

func TestTemplate_Build(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	data := `
	{
		"name": "my-image",
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	expectedConfig := map[string]interface{}{
		"name": "test1",
		"type": "test-builder",
	}

	template, err := ParseTemplate([]byte(data))
	assert.Nil(err, "should not error")

	builder := testBuilder()
	builderMap := map[string]Builder{
		"test-builder": builder,
	}

	builderFactory := func(n string) (Builder, error) { return builderMap[n], nil }
	components := &ComponentFinder{Builder: builderFactory}

	// Get the build, verifying we can get it without issue, but also
	// that the proper builder was looked up and used for the build.
	build, err := template.Build("test1", components)
	assert.Nil(err, "should not error")

	build.Prepare()
	build.Run(testUi())

	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, expectedConfig, "prepare config should be correct")
	assert.True(builder.runCalled, "run should be called")
}
