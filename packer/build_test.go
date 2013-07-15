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
			"foo": []Hook{&TestHook{}},
		},
		provisioners: []coreBuildProvisioner{
			coreBuildProvisioner{&TestProvisioner{}, []interface{}{42}},
		},
		postProcessors: [][]coreBuildPostProcessor{
			[]coreBuildPostProcessor{
				coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp"}, "testPP", 42, true},
			},
		},
	}
}

func testBuilder() *TestBuilder {
	return &TestBuilder{}
}

func TestBuild_Name(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	assert.Equal(build.Name(), "test", "should have a name")
}

func TestBuild_Prepare(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	packerConfig := map[string]interface{}{
		BuildNameConfigKey:   "test",
		BuilderTypeConfigKey: "foo",
		DebugConfigKey:       false,
		ForceConfigKey:       false,
	}

	build := testBuild()
	builder := build.builder.(*TestBuilder)

	build.Prepare()
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, packerConfig}, "prepare config should be 42")

	coreProv := build.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfigs, []interface{}{42, packerConfig}, "prepare should be called with proper config")

	corePP := build.postProcessors[0][0]
	pp := corePP.processor.(*TestPostProcessor)
	assert.True(pp.configCalled, "config should be called")
	assert.Equal(pp.configVal, []interface{}{42, packerConfig}, "config should have right value")
}

func TestBuild_Prepare_Twice(t *testing.T) {
	build := testBuild()
	if err := build.Prepare(); err != nil {
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

	build.Prepare()
}

func TestBuild_Prepare_Debug(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	packerConfig := map[string]interface{}{
		BuildNameConfigKey:   "test",
		BuilderTypeConfigKey: "foo",
		DebugConfigKey:       true,
		ForceConfigKey:       false,
	}

	build := testBuild()
	builder := build.builder.(*TestBuilder)

	build.SetDebug(true)
	build.Prepare()
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, packerConfig}, "prepare config should be 42")

	coreProv := build.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfigs, []interface{}{42, packerConfig}, "prepare should be called with proper config")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cache := &TestCache{}
	ui := testUi()

	build := testBuild()
	build.Prepare()
	artifacts, err := build.Run(ui, cache)
	assert.Nil(err, "should not error")
	assert.Equal(len(artifacts), 2, "should have two artifacts")

	// Verify builder was run
	builder := build.builder.(*TestBuilder)
	assert.True(builder.runCalled, "run should be called")

	// Verify hooks are disapatchable
	dispatchHook := builder.runHook
	dispatchHook.Run("foo", nil, nil, 42)

	hook := build.hooks["foo"][0].(*TestHook)
	assert.True(hook.runCalled, "run should be called")
	assert.Equal(hook.runData, 42, "should have correct data")

	// Verify provisioners run
	dispatchHook.Run(HookProvision, nil, nil, 42)
	prov := build.provisioners[0].provisioner.(*TestProvisioner)
	assert.True(prov.provCalled, "provision should be called")

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

	build.Prepare()
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
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp"}, "pp", 42, false},
		},
	}

	build.Prepare()
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
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1"}, "pp", 42, false},
		},
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2"}, "pp", 42, true},
		},
	}

	build.Prepare()
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
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1a"}, "pp", 42, false},
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp1b"}, "pp", 42, true},
		},
		[]coreBuildPostProcessor{
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2a"}, "pp", 42, false},
			coreBuildPostProcessor{&TestPostProcessor{artifactId: "pp2b"}, "pp", 42, false},
		},
	}

	build.Prepare()
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
				&TestPostProcessor{artifactId: "pp", keep: true}, "pp", 42, false,
			},
		},
	}

	build.Prepare()
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
