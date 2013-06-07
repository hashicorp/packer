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

	build := testBuild()
	ui := testUi()

	coreB := build.(*coreBuild)
	builder := coreB.builder.(*TestBuilder)

	build.Prepare(ui)
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, 42, "prepare config should be 42")

	// Verify provisioners were prepared
	coreProv := coreB.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfigs, []interface{}{42}, "prepare should be called with proper config")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := testUi()

	build := testBuild()
	build.Prepare(ui)
	build.Run(ui)

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

	testBuild().Run(testUi())
}

func TestBuild_Cancel(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	build.Cancel()

	coreB := build.(*coreBuild)

	builder := coreB.builder.(*TestBuilder)
	assert.True(builder.cancelCalled, "cancel should be called")
}
