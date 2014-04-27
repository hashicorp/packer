package packer

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"
)

func testTemplateComponentFinder() *ComponentFinder {
	builder := new(MockBuilder)
	pp := new(TestPostProcessor)
	provisioner := &MockProvisioner{}

	builderMap := map[string]Builder{
		"test-builder": builder,
	}

	ppMap := map[string]PostProcessor{
		"test-pp": pp,
	}

	provisionerMap := map[string]Provisioner{
		"test-prov": provisioner,
	}

	builderFactory := func(n string) (Builder, error) { return builderMap[n], nil }
	ppFactory := func(n string) (PostProcessor, error) { return ppMap[n], nil }
	provFactory := func(n string) (Provisioner, error) { return provisionerMap[n], nil }
	return &ComponentFinder{
		Builder:       builderFactory,
		PostProcessor: ppFactory,
		Provisioner:   provFactory,
	}
}

func TestParseTemplateFile_basic(t *testing.T) {
	data := `
	{
		"builders": [{"type": "something"}]
	}
	`

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(data))
	tf.Close()

	result, err := ParseTemplateFile(tf.Name(), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result.Builders) != 1 {
		t.Fatalf("bad: %#v", result.Builders)
	}
}

func TestParseTemplateFile_minPackerVersionBad(t *testing.T) {
	data := `
	{
		"min_packer_version": "27.0.0",
		"builders": [{"type": "something"}]
	}
	`

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(data))
	tf.Close()

	_, err = ParseTemplateFile(tf.Name(), nil)
	if err == nil {
		t.Fatal("expects error")
	}
}

func TestParseTemplateFile_minPackerVersionFormat(t *testing.T) {
	data := `
	{
		"min_packer_version": "NOPE NOPE NOPE",
		"builders": [{"type": "something"}]
	}
	`

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(data))
	tf.Close()

	_, err = ParseTemplateFile(tf.Name(), nil)
	if err == nil {
		t.Fatal("expects error")
	}
}

func TestParseTemplateFile_minPackerVersionGood(t *testing.T) {
	data := `
	{
		"min_packer_version": "0.1",
		"builders": [{"type": "something"}]
	}
	`

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Write([]byte(data))
	tf.Close()

	_, err = ParseTemplateFile(tf.Name(), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestParseTemplateFile_stdin(t *testing.T) {
	data := `
	{
		"builders": [{"type": "something"}]
	}
	`

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()
	tf.Write([]byte(data))

	// Sync and seek to the beginning so that we can re-read the contents
	tf.Sync()
	tf.Seek(0, 0)

	// Set stdin to something we control
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()
	os.Stdin = tf

	result, err := ParseTemplateFile("-", nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(result.Builders) != 1 {
		t.Fatalf("bad: %#v", result.Builders)
	}
}

func TestParseTemplate_Basic(t *testing.T) {
	data := `
	{
		"builders": [{"type": "something"}]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Builders) != 1 {
		t.Fatalf("bad: %#v", result.Builders)
	}
}

func TestParseTemplate_Description(t *testing.T) {
	data := `
	{
		"description": "Foo",
		"builders": [{"type": "something"}]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if result.Description != "Foo" {
		t.Fatalf("bad: %#v", result.Description)
	}
}

func TestParseTemplate_Invalid(t *testing.T) {
	// Note there is an extra comma below for a purposeful
	// syntax error in the JSON.
	data := `
	{
		"builders": [],
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("shold have error")
	}
	if result != nil {
		t.Fatal("should not have result")
	}
}

func TestParseTemplate_InvalidKeys(t *testing.T) {
	// Note there is an extra comma below for a purposeful
	// syntax error in the JSON.
	data := `
	{
		"builders": [{"type": "foo"}],
		"what is this": ""
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
	if result != nil {
		t.Fatal("should not have result")
	}
}

func TestParseTemplate_BuilderWithoutType(t *testing.T) {
	data := `
	{
		"builders": [{}]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestParseTemplate_BuilderWithNonStringType(t *testing.T) {
	data := `
	{
		"builders": [{
			"type": 42
		}]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestParseTemplate_BuilderWithoutName(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"type": "amazon-ebs"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Builders) != 1 {
		t.Fatalf("bad: %#v", result.Builders)
	}

	builder, ok := result.Builders["amazon-ebs"]
	if !ok {
		t.Fatal("should be ok")
	}
	if builder.Type != "amazon-ebs" {
		t.Fatalf("bad: %#v", builder.Type)
	}
}

func TestParseTemplate_BuilderWithName(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "bob",
				"type": "amazon-ebs"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Builders) != 1 {
		t.Fatalf("bad: %#v", result.Builders)
	}

	builder, ok := result.Builders["bob"]
	if !ok {
		t.Fatal("should be ok")
	}
	if builder.Type != "amazon-ebs" {
		t.Fatalf("bad: %#v", builder.Type)
	}

	RawConfig := builder.RawConfig
	if RawConfig == nil {
		t.Fatal("missing builder raw config")
	}

	expected := map[string]interface{}{
		"type": "amazon-ebs",
	}

	if !reflect.DeepEqual(RawConfig, expected) {
		t.Fatalf("bad raw: %#v", RawConfig)
	}
}

func TestParseTemplate_BuilderWithConflictingName(t *testing.T) {
	data := `
	{
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

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestParseTemplate_Hooks(t *testing.T) {
	data := `
	{

		"builders": [{"type": "foo"}],

		"hooks": {
			"event": ["foo", "bar"]
		}
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Hooks) != 1 {
		t.Fatalf("bad: %#v", result.Hooks)
	}

	hooks, ok := result.Hooks["event"]
	if !ok {
		t.Fatal("should be okay")
	}
	if !reflect.DeepEqual(hooks, []string{"foo", "bar"}) {
		t.Fatalf("bad: %#v", hooks)
	}
}

func TestParseTemplate_PostProcessors(t *testing.T) {
	data := `
	{
		"builders": [{"type": "foo"}],

		"post-processors": [
			"simple",

			{ "type": "detailed" },

			[ "foo", { "type": "bar" } ]
		]
	}
	`

	tpl, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("error parsing: %s", err)
	}

	if len(tpl.PostProcessors) != 3 {
		t.Fatalf("bad number of post-processors: %d", len(tpl.PostProcessors))
	}

	pp := tpl.PostProcessors[0]
	if len(pp) != 1 {
		t.Fatalf("wrong number of configs in simple: %d", len(pp))
	}

	if pp[0].Type != "simple" {
		t.Fatalf("wrong type for simple: %s", pp[0].Type)
	}

	pp = tpl.PostProcessors[1]
	if len(pp) != 1 {
		t.Fatalf("wrong number of configs in detailed: %d", len(pp))
	}

	if pp[0].Type != "detailed" {
		t.Fatalf("wrong type for detailed: %s", pp[0].Type)
	}

	pp = tpl.PostProcessors[2]
	if len(pp) != 2 {
		t.Fatalf("wrong number of configs for sequence: %d", len(pp))
	}

	if pp[0].Type != "foo" {
		t.Fatalf("wrong type for sequence 0: %s", pp[0].Type)
	}

	if pp[1].Type != "bar" {
		t.Fatalf("wrong type for sequence 1: %s", pp[1].Type)
	}
}

func TestParseTemplate_ProvisionerWithoutType(t *testing.T) {
	data := `
	{
		"builders": [{"type": "foo"}],

		"provisioners": [{}]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestParseTemplate_ProvisionerWithNonStringType(t *testing.T) {
	data := `
	{
		"builders": [{"type": "foo"}],

		"provisioners": [{
			"type": 42
		}]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestParseTemplate_Provisioners(t *testing.T) {
	data := `
	{
		"builders": [{"type": "foo"}],

		"provisioners": [
			{
				"type": "shell"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatal("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Provisioners) != 1 {
		t.Fatalf("bad: %#v", result.Provisioners)
	}
	if result.Provisioners[0].Type != "shell" {
		t.Fatalf("bad: %#v", result.Provisioners[0].Type)
	}
	if result.Provisioners[0].RawConfig == nil {
		t.Fatal("should have raw config")
	}
}

func TestParseTemplate_ProvisionerPauseBefore(t *testing.T) {
	data := `
	{
		"builders": [{"type": "foo"}],

		"provisioners": [
			{
				"type": "shell",
				"pause_before": "10s"
			}
		]
	}
	`

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatal("err: %s", err)
	}
	if result == nil {
		t.Fatal("should have result")
	}
	if len(result.Provisioners) != 1 {
		t.Fatalf("bad: %#v", result.Provisioners)
	}
	if result.Provisioners[0].Type != "shell" {
		t.Fatalf("bad: %#v", result.Provisioners[0].Type)
	}
	if result.Provisioners[0].pauseBefore != 10*time.Second {
		t.Fatalf("bad: %s", result.Provisioners[0].pauseBefore)
	}
}

func TestParseTemplate_Variables(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": "bar",
			"bar": null,
			"baz": 27
		},

		"builders": [{"type": "something"}]
	}
	`

	result, err := ParseTemplate([]byte(data), map[string]string{
		"bar": "bar",
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.Variables == nil || len(result.Variables) != 3 {
		t.Fatalf("bad vars: %#v", result.Variables)
	}

	if result.Variables["foo"].Default != "bar" {
		t.Fatal("foo default is not right")
	}
	if result.Variables["foo"].Required {
		t.Fatal("foo should not be required")
	}
	if result.Variables["foo"].HasValue {
		t.Fatal("foo should not have value")
	}

	if result.Variables["bar"].Default != "" {
		t.Fatal("default should be empty")
	}
	if !result.Variables["bar"].Required {
		t.Fatal("bar should be required")
	}
	if !result.Variables["bar"].HasValue {
		t.Fatal("bar should have value")
	}
	if result.Variables["bar"].Value != "bar" {
		t.Fatal("bad value")
	}

	if result.Variables["baz"].Default != "27" {
		t.Fatal("default should be empty")
	}

	if result.Variables["baz"].Required {
		t.Fatal("baz should not be required")
	}
}

func TestParseTemplate_variablesSet(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": "bar"
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), map[string]string{
		"foo": "value",
	})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(template.Variables) != 1 {
		t.Fatalf("bad vars: %#v", template.Variables)
	}
	if template.Variables["foo"].Value != "value" {
		t.Fatalf("bad: %#v", template.Variables["foo"])
	}
}

func TestParseTemplate_variablesSetUnknown(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": "bar"
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), map[string]string{
		"what": "value",
	})
	if err == nil {
		t.Fatal("should error")
	}
}

func TestParseTemplate_variablesBadDefault(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": 7,
		},

		"builders": [{"type": "something"}]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplate_BuildNames(t *testing.T) {
	data := `
	{
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

	result, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	buildNames := result.BuildNames()
	sort.Strings(buildNames)
	if !reflect.DeepEqual(buildNames, []string{"bob", "chris"}) {
		t.Fatalf("bad: %#v", buildNames)
	}
}

func TestTemplate_BuildUnknown(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	build, err := template.Build("nope", nil)
	if build != nil {
		t.Fatalf("build should be nil: %#v", build)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplate_BuildUnknownBuilder(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	builderFactory := func(string) (Builder, error) { return nil, nil }
	components := &ComponentFinder{Builder: builderFactory}
	build, err := template.Build("test1", components)
	if err == nil {
		t.Fatal("should have error")
	}
	if build != nil {
		t.Fatalf("bad: %#v", build)
	}
}

func TestTemplateBuild_envInVars(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": "{{env \"foo\"}}"
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	defer os.Setenv("foo", os.Getenv("foo"))
	if err := os.Setenv("foo", "bar"); err != nil {
		t.Fatalf("err: %s", err)
	}

	template, err := ParseTemplate([]byte(data), map[string]string{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	b, err := template.Build("test1", testComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	coreBuild, ok := b.(*coreBuild)
	if !ok {
		t.Fatal("should be ok")
	}

	if coreBuild.variables["foo"] != "bar" {
		t.Fatalf("bad: %#v", coreBuild.variables)
	}
}

func TestTemplateBuild_names(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": null
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2-{{user \"foo\"}}",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	b, err := template.Build("test1", testComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if b.Name() != "test1" {
		t.Fatalf("bad: %#v", b.Name())
	}

	b, err = template.Build("test2-{{user \"foo\"}}", testComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if b.Name() != "test2-bar" {
		t.Fatalf("bad: %#v", b.Name())
	}
}

func TestTemplate_Build_NilBuilderFunc(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer func() {
		p := recover()
		if p == nil {
			t.Fatal("should panic")
		}

		if p.(string) != "no builder function" {
			t.Fatalf("bad panic: %s", p.(string))
		}
	}()

	template.Build("test1", &ComponentFinder{})
}

func TestTemplate_Build_NilProvisionerFunc(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer func() {
		p := recover()
		if p == nil {
			t.Fatal("should panic")
		}

		if p.(string) != "no provisioner function" {
			t.Fatalf("bad panic: %s", p.(string))
		}
	}()

	template.Build("test1", &ComponentFinder{
		Builder: func(string) (Builder, error) { return nil, nil },
	})
}

func TestTemplate_Build_NilProvisionerFunc_WithNoProvisioners(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": []
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	template.Build("test1", &ComponentFinder{
		Builder: func(string) (Builder, error) { return nil, nil },
	})
}

func TestTemplate_Build(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov"
			}
		],

		"post-processors": [
			"simple",
			[
				"simple",
				{ "type": "simple", "keep_input_artifact": true }
			]
		]
	}
	`

	expectedConfig := map[string]interface{}{
		"type": "test-builder",
	}

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	builder := new(MockBuilder)
	builderMap := map[string]Builder{
		"test-builder": builder,
	}

	provisioner := &MockProvisioner{}
	provisionerMap := map[string]Provisioner{
		"test-prov": provisioner,
	}

	pp := new(TestPostProcessor)
	ppMap := map[string]PostProcessor{
		"simple": pp,
	}

	builderFactory := func(n string) (Builder, error) { return builderMap[n], nil }
	ppFactory := func(n string) (PostProcessor, error) { return ppMap[n], nil }
	provFactory := func(n string) (Provisioner, error) { return provisionerMap[n], nil }
	components := &ComponentFinder{
		Builder:       builderFactory,
		PostProcessor: ppFactory,
		Provisioner:   provFactory,
	}

	// Get the build, verifying we can get it without issue, but also
	// that the proper builder was looked up and used for the build.
	build, err := template.Build("test1", components)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	coreBuild, ok := build.(*coreBuild)
	if !ok {
		t.Fatal("should be ok")
	}
	if coreBuild.builder != builder {
		t.Fatalf("bad: %#v", coreBuild.builder)
	}
	if !reflect.DeepEqual(coreBuild.builderConfig, expectedConfig) {
		t.Fatalf("bad: %#v", coreBuild.builderConfig)
	}
	if len(coreBuild.provisioners) != 1 {
		t.Fatalf("bad: %#v", coreBuild.provisioners)
	}
	if len(coreBuild.postProcessors) != 2 {
		t.Fatalf("bad: %#v", coreBuild.postProcessors)
	}

	if len(coreBuild.postProcessors[0]) != 1 {
		t.Fatalf("bad: %#v", coreBuild.postProcessors[0])
	}
	if len(coreBuild.postProcessors[1]) != 2 {
		t.Fatalf("bad: %#v", coreBuild.postProcessors[1])
	}

	if coreBuild.postProcessors[1][0].keepInputArtifact {
		t.Fatal("postProcessors[1][0] should not keep input artifact")
	}
	if !coreBuild.postProcessors[1][1].keepInputArtifact {
		t.Fatal("postProcessors[1][1] should keep input artifact")
	}

	config := coreBuild.postProcessors[1][1].config
	if _, ok := config["keep_input_artifact"]; ok {
		t.Fatal("should not have keep_input_artifact")
	}
}

func TestTemplateBuild_exceptOnlyPP(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"post-processors": [
			{
				"type": "test-pp",
				"except": ["test1"],
				"only": ["test1"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_exceptOnlyProv(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"except": ["test1"],
				"only": ["test1"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_exceptPPInvalid(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"post-processors": [
			{
				"type": "test-pp",
				"except": ["test5"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_exceptPP(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"post-processors": [
			{
				"type": "test-pp",
				"except": ["test1"]
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify test1 has no post-processors
	build, err := template.Build("test1", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild := build.(*coreBuild)
	if len(cbuild.postProcessors) > 0 {
		t.Fatal("should have no postProcessors")
	}

	// Verify test2 has no post-processors
	build, err = template.Build("test2", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild = build.(*coreBuild)
	if len(cbuild.postProcessors) != 1 {
		t.Fatalf("invalid: %d", len(cbuild.postProcessors))
	}
}

func TestTemplateBuild_exceptProvInvalid(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"except": ["test5"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_exceptProv(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"except": ["test1"]
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify test1 has no provisioners
	build, err := template.Build("test1", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild := build.(*coreBuild)
	if len(cbuild.provisioners) > 0 {
		t.Fatal("should have no provisioners")
	}

	// Verify test2 has no provisioners
	build, err = template.Build("test2", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild = build.(*coreBuild)
	if len(cbuild.provisioners) != 1 {
		t.Fatalf("invalid: %d", len(cbuild.provisioners))
	}
}

func TestTemplateBuild_onlyPPInvalid(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"post-processors": [
			{
				"type": "test-pp",
				"only": ["test5"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_onlyPP(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"post-processors": [
			{
				"type": "test-pp",
				"only": ["test2"]
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify test1 has no post-processors
	build, err := template.Build("test1", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild := build.(*coreBuild)
	if len(cbuild.postProcessors) > 0 {
		t.Fatal("should have no postProcessors")
	}

	// Verify test2 has no post-processors
	build, err = template.Build("test2", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild = build.(*coreBuild)
	if len(cbuild.postProcessors) != 1 {
		t.Fatalf("invalid: %d", len(cbuild.postProcessors))
	}
}

func TestTemplateBuild_onlyProvInvalid(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"only": ["test5"]
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_onlyProv(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			},
			{
				"name": "test2",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"only": ["test2"]
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify test1 has no provisioners
	build, err := template.Build("test1", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild := build.(*coreBuild)
	if len(cbuild.provisioners) > 0 {
		t.Fatal("should have no provisioners")
	}

	// Verify test2 has no provisioners
	build, err = template.Build("test2", testTemplateComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cbuild = build.(*coreBuild)
	if len(cbuild.provisioners) != 1 {
		t.Fatalf("invalid: %d", len(cbuild.provisioners))
	}
}

func TestTemplate_Build_ProvisionerOverride(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",

				"override": {
					"test1": {}
				}
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	RawConfig := template.Provisioners[0].RawConfig
	if RawConfig == nil {
		t.Fatal("missing provisioner raw config")
	}

	expected := map[string]interface{}{
		"type": "test-prov",
	}

	if !reflect.DeepEqual(RawConfig, expected) {
		t.Fatalf("bad raw: %#v", RawConfig)
	}

	builder := new(MockBuilder)
	builderMap := map[string]Builder{
		"test-builder": builder,
	}

	provisioner := &MockProvisioner{}
	provisionerMap := map[string]Provisioner{
		"test-prov": provisioner,
	}

	builderFactory := func(n string) (Builder, error) { return builderMap[n], nil }
	provFactory := func(n string) (Provisioner, error) { return provisionerMap[n], nil }
	components := &ComponentFinder{
		Builder:     builderFactory,
		Provisioner: provFactory,
	}

	// Get the build, verifying we can get it without issue, but also
	// that the proper builder was looked up and used for the build.
	build, err := template.Build("test1", components)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	coreBuild, ok := build.(*coreBuild)
	if !ok {
		t.Fatal("should be okay")
	}
	if len(coreBuild.provisioners) != 1 {
		t.Fatalf("bad: %#v", coreBuild.provisioners)
	}
	if len(coreBuild.provisioners[0].config) != 2 {
		t.Fatalf("bad: %#v", coreBuild.provisioners[0].config)
	}
}

func TestTemplate_Build_ProvisionerOverrideBad(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",

				"override": {
					"testNope": {}
				}
			}
		]
	}
	`

	_, err := ParseTemplate([]byte(data), nil)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestTemplateBuild_ProvisionerPauseBefore(t *testing.T) {
	data := `
	{
		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		],

		"provisioners": [
			{
				"type": "test-prov",
				"pause_before": "5s"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	builder := new(MockBuilder)
	builderMap := map[string]Builder{
		"test-builder": builder,
	}

	provisioner := &MockProvisioner{}
	provisionerMap := map[string]Provisioner{
		"test-prov": provisioner,
	}

	builderFactory := func(n string) (Builder, error) { return builderMap[n], nil }
	provFactory := func(n string) (Provisioner, error) { return provisionerMap[n], nil }
	components := &ComponentFinder{
		Builder:     builderFactory,
		Provisioner: provFactory,
	}

	// Get the build, verifying we can get it without issue, but also
	// that the proper builder was looked up and used for the build.
	build, err := template.Build("test1", components)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	coreBuild, ok := build.(*coreBuild)
	if !ok {
		t.Fatal("should be okay")
	}
	if len(coreBuild.provisioners) != 1 {
		t.Fatalf("bad: %#v", coreBuild.provisioners)
	}
	if pp, ok := coreBuild.provisioners[0].provisioner.(*PausedProvisioner); !ok {
		t.Fatalf("should be paused provisioner")
	} else {
		if pp.PauseBefore != 5*time.Second {
			t.Fatalf("bad: %#v", pp.PauseBefore)
		}
	}

	config := coreBuild.provisioners[0].config[0].(map[string]interface{})
	if _, ok := config["pause_before"]; ok {
		t.Fatal("pause_before should be removed")
	}
}

func TestTemplateBuild_variables(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": "bar"
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	build, err := template.Build("test1", testComponentFinder())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	coreBuild, ok := build.(*coreBuild)
	if !ok {
		t.Fatalf("couldn't convert!")
	}

	expected := map[string]string{"foo": "bar"}
	if !reflect.DeepEqual(coreBuild.variables, expected) {
		t.Fatalf("bad vars: %#v", coreBuild.variables)
	}
}

func TestTemplateBuild_variablesRequiredNotSet(t *testing.T) {
	data := `
	{
		"variables": {
			"foo": null
		},

		"builders": [
			{
				"name": "test1",
				"type": "test-builder"
			}
		]
	}
	`

	template, err := ParseTemplate([]byte(data), map[string]string{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	_, err = template.Build("test1", testComponentFinder())
	if err == nil {
		t.Fatal("should error")
	}
}
