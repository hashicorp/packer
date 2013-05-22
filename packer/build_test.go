package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

type TestBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled     bool
	runHook       Hook
	runUi         Ui
}

func (tb *TestBuilder) Prepare(config interface{}) error {
	tb.prepareCalled = true
	tb.prepareConfig = config
	return nil
}

func (tb *TestBuilder) Run(ui Ui, h Hook) Artifact {
	tb.runCalled = true
	tb.runHook = h
	tb.runUi = ui
	return nil
}

func testBuild() Build {
	return &coreBuild{
		name:          "test",
		builder:       &TestBuilder{},
		builderConfig: 42,
		provisioners: []coreBuildProvisioner{
			coreBuildProvisioner{&TestProvisioner{}, 42},
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
	builder := build.(*coreBuild).builder.(*TestBuilder)

	build.Prepare(nil)
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, 42, "prepare config should be 42")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := testUi()

	build := testBuild()
	build.Prepare(ui)
	build.Run(ui)

	coreB := build.(*coreBuild)

	// Verify builder was prepared
	builder := coreB.builder.(*TestBuilder)
	assert.True(builder.runCalled, "run should be called")
	assert.Equal(builder.runUi, ui, "run should be called with ui")

	// Verify provisioners were prepared
	coreProv := coreB.provisioners[0]
	prov := coreProv.provisioner.(*TestProvisioner)
	assert.True(prov.prepCalled, "prepare should be called")
	assert.Equal(prov.prepConfig, 42, "prepare should be called with proper config")
	assert.Equal(prov.prepUi, ui, "prepare should be called with proper ui")
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
