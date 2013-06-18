package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func testBuild() Build {
	return &coreBuild{
		name:          "test",
		builder:       &TestBuilder{},
		builderConfig: 42,
		hooks: map[string][]Hook{
			"foo": []Hook{&TestHook{}},
		},
		provisioners: []coreBuildProvisioner{
			coreBuildProvisioner{&TestProvisioner{}, []interface{}{42}},
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

	debugFalseConfig := map[string]interface{}{DebugConfigKey: false}

	build := testBuild()
	coreB := build.(*coreBuild)
	builder := coreB.builder.(*TestBuilder)

	build.Prepare()
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, debugFalseConfig}, "prepare config should be 42")

	coreProv := coreB.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfigs, []interface{}{42, debugFalseConfig}, "prepare should be called with proper config")
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

	debugConfig := map[string]interface{}{DebugConfigKey: true}

	build := testBuild()
	coreB := build.(*coreBuild)
	builder := coreB.builder.(*TestBuilder)

	build.SetDebug(true)
	build.Prepare()
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, []interface{}{42, debugConfig}, "prepare config should be 42")

	coreProv := coreB.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfigs, []interface{}{42, debugConfig}, "prepare should be called with proper config")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	cache := &TestCache{}
	ui := testUi()

	build := testBuild()
	build.Prepare()
	build.Run(ui, cache)

	coreB := build.(*coreBuild)

	// Verify builder was run
	builder := coreB.builder.(*TestBuilder)
	assert.True(builder.runCalled, "run should be called")
	assert.Equal(builder.runUi, ui, "run should be called with ui")

	// Verify hooks are disapatchable
	dispatchHook := builder.runHook
	dispatchHook.Run("foo", nil, nil, 42)

	hook := coreB.hooks["foo"][0].(*TestHook)
	assert.True(hook.runCalled, "run should be called")
	assert.Equal(hook.runData, 42, "should have correct data")

	// Verify provisioners run
	dispatchHook.Run(HookProvision, nil, nil, 42)
	prov := coreB.provisioners[0].provisioner.(*TestProvisioner)
	assert.True(prov.provCalled, "provision should be called")
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

	coreB := build.(*coreBuild)

	builder := coreB.builder.(*TestBuilder)
	assert.True(builder.cancelCalled, "cancel should be called")
}
