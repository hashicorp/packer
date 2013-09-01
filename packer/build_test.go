package packer

import (
	"cgl.tideland.biz/asserts"
	"reflect"
	"testing"
)

func testBuild() *coreBuild {
	return &coreBuild{
		name:          "test",
		builder:       &TestBuilder{artifactId: "b"},
		builderConfig: 42,
		builderType:   "foo",
		hooks: map[string][]Hook{
			"foo": []Hook{&MockHook{}},
		},
		provisioners: []coreBuildProvisioner{
			coreBuildProvisioner{&MockProvisioner{}, []interface{}{42}},
		},
		postProcessors: [][]coreBuildPostProcessor{
			[]coreBuildPostProcessor{
				coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp"}, "testPP", make(map[string]interface{}), true},
			},
		},
		variables: make(map[string]coreBuildVariable),
	}
}

func testBuilder() *TestBuilder {
	return &TestBuilder{}
}

func testDefaultPackerConfig() map[string]interface{} {
	return map[string]interface{}{
		BuildNameConfigKey:     "test",
		BuilderTypeConfigKey:   "foo",
		DebugConfigKey:         false,
		ForceConfigKey:         false,
		UserVariablesConfigKey: make(map[string]string),
	}
}
func TestBuild_Name(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	assert.Equal(build.Name(), "test", "should have a name")
}

func TestBuild_Prepare(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	packerConfig := testDefaultPackerConfig()

	build := testBuild()
	builder := build.builder.(*TestBuilder)

	build.Prepare(nil)
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, packerConfig}, "prepare config should be 42")

	coreProv := build.provisioners[0]
	prov := coreProv.provisioner.(*MockProvisioner)
	assert.True(prov.PrepCalled, "prepare should be called")
	assert.Equal(prov.PrepConfigs, []interface{}{42, packerConfig}, "prepare should be called with proper config")

	corePP := build.postProcessors[0][0]
	pp := corePP.processor.(*TestPostProcessor)
	assert.True(pp.configCalled, "config should be called")
	assert.Equal(pp.configVal, []interface{}{make(map[string]interface{}), packerConfig}, "config should have right value")
}

func TestBuild_Prepare_Twice(t *testing.T) {
	build := testBuild()
	if err := build.Prepare(nil); err != nil {
		t.Fatalf("bad error: %s", err)
	}

	defer func() {
		p := recover()
		if p == nil {
			t.Fatalf("should've paniced")
		}

		if p.(string) != "prepare already called" {
			t.Fatalf("Invalid panic: %s", p)
		}
	}()

	build.Prepare(nil)
}

func TestBuild_Prepare_Debug(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	packerConfig := testDefaultPackerConfig()
	packerConfig[DebugConfigKey] = true

	build := testBuild()
	builder := build.builder.(*TestBuilder)

	build.SetDebug(true)
	build.Prepare(nil)
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, packerConfig}, "prepare config should be 42")

	coreProv := build.provisioners[0]
	prov := coreProv.provisioner.(*MockProvisioner)
	assert.True(prov.PrepCalled, "prepare should be called")
	assert.Equal(prov.PrepConfigs, []interface{}{42, packerConfig}, "prepare should be called with proper config")
}

func TestBuildPrepare_variables_default(t *testing.T) {
	packerConfig := testDefaultPackerConfig()
	packerConfig[UserVariablesConfigKey] = map[string]string{
		"foo": "bar",
	}

	build := testBuild()
	build.variables["foo"] = coreBuildVariable{Default: "bar"}
	builder := build.builder.(*TestBuilder)

	err := build.Prepare(nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !builder.prepareCalled {
		t.Fatal("prepare should be called")
	}

	if !reflect.DeepEqual(builder.prepareConfig[1], packerConfig) {
		t.Fatalf("prepare bad: %#v", builder.prepareConfig[1])
	}
}

func TestBuildPrepare_variables_nonexist(t *testing.T) {
	build := testBuild()
	build.variables["foo"] = coreBuildVariable{Default: "bar"}

	err := build.Prepare(map[string]string{"bar": "baz"})
	if err == nil {
		t.Fatal("should have had error")
	}
}

func TestBuildPrepare_variables_override(t *testing.T) {
	packerConfig := testDefaultPackerConfig()
	packerConfig[UserVariablesConfigKey] = map[string]string{
		"foo": "baz",
	}

	build := testBuild()
	build.variables["foo"] = coreBuildVariable{Default: "bar"}
	builder := build.builder.(*TestBuilder)

	err := build.Prepare(map[string]string{"foo": "baz"})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !builder.prepareCalled {
		t.Fatal("prepare should be called")
	}

	if !reflect.DeepEqual(builder.prepareConfig[1], packerConfig) {
		t.Fatalf("prepare bad: %#v", builder.prepareConfig[1])
	}
}

func TestBuildPrepare_variablesRequired(t *testing.T) {
	build := testBuild()
	build.variables["foo"] = coreBuildVariable{Required: true}

	err := build.Prepare(map[string]string{})
	if err == nil {
		t.Fatal("should have had error")
	}

	// Test with setting the value
	build = testBuild()
	build.variables["foo"] = coreBuildVariable{Required: true}
	err = build.Prepare(map[string]string{"foo": ""})
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cache := &TestCache{}
	ui := testUi()

	build := testBuild()
	build.Prepare(nil)
	artifacts, err := build.Run(ui, cache)
	assert.Nil(err, "should not error")
	assert.Equal(len(artifacts), 2, "should have two artifacts")

	// Verify builder was run
	builder := build.builder.(*TestBuilder)
	assert.True(builder.runCalled, "run should be called")

	// Verify hooks are disapatchable
	dispatchHook := builder.runHook
	dispatchHook.Run("foo", nil, nil, 42)

	hook := build.hooks["foo"][0].(*MockHook)
	assert.True(hook.RunCalled, "run should be called")
	assert.Equal(hook.RunData, 42, "should have correct data")

	// Verify provisioners run
	dispatchHook.Run(HookProvision, nil, nil, 42)
	prov := build.provisioners[0].provisioner.(*MockProvisioner)
	assert.True(prov.ProvCalled, "provision should be called")

	// Verify post-processor was run
	pp := build.postProcessors[0][0].processor.(*TestPostProcessor)
	assert.True(pp.ppCalled, "post processor should be called")
}

func TestBuild_Run_Artifacts(t *testing.T) {
	cache := &TestCache{}
	ui := testUi()

	// Test case: Test that with no post-processors, we only get the
	// main build.
	build := testBuild()
	build.postProcessors = [][]coreBuildPostProcessor{}

	build.Prepare(nil)
	artifacts, err := build.Run(ui, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds := []string{"b"}
	artifactIds := make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that doesn't keep
	// inputs, only that post-processors results are returned.
	build = testBuild()
	build.postProcessors = [][]coreBuildPostProcessor{
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp"}, "pp", make(map[string]interface{}), false},
		},
	}

	build.Prepare(nil)
	artifacts, err = build.Run(ui, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with multiple post-processors, as long as one
	// keeps the original, the original is kept.
	build = testBuild()
	build.postProcessors = [][]coreBuildPostProcessor{
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1"}, "pp", make(map[string]interface{}), false},
		},
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2"}, "pp", make(map[string]interface{}), true},
		},
	}

	build.Prepare(nil)
	artifacts, err = build.Run(ui, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"b", "pp1", "pp2"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with sequences, intermediaries are kept if they
	// want to be.
	build = testBuild()
	build.postProcessors = [][]coreBuildPostProcessor{
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1a"}, "pp", make(map[string]interface{}), false},
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1b"}, "pp", make(map[string]interface{}), true},
		},
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2a"}, "pp", make(map[string]interface{}), false},
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2b"}, "pp", make(map[string]interface{}), false},
		},
	}

	build.Prepare(nil)
	artifacts, err = build.Run(ui, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"pp1a", "pp1b", "pp2b"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}

	// Test case: Test that with a single post-processor that forcibly
	// keeps inputs, that the artifacts are kept.
	build = testBuild()
	build.postProcessors = [][]coreBuildPostProcessor{
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{
				&TestPostProcessor{artifactId: "pp", keep: true}, "pp", make(map[string]interface{}), false,
			},
		},
	}

	build.Prepare(nil)
	artifacts, err = build.Run(ui, cache)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedIds = []string{"b", "pp"}
	artifactIds = make([]string, len(artifacts))
	for i, artifact := range artifacts {
		artifactIds[i] = artifact.Id()
	}

	if !reflect.DeepEqual(artifactIds, expectedIds) {
		t.Fatalf("unexpected ids: %#v", artifactIds)
	}
}

func TestBuild_RunBeforePrepare(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defer func() {
		p := recover()
		assert.NotNil(p, "should panic")
		assert.Equal(p.(string), "Prepare must be called first", "right panic")
	}()

	testBuild().Run(testUi(), &TestCache{})
}

func TestBuild_Cancel(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	build.Cancel()

	builder := build.builder.(*TestBuilder)
	assert.True(builder.cancelCalled, "cancel should be called")
}
