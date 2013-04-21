package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

type TestBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled bool
	runBuild *Build
	runUi Ui
}

func (tb *TestBuilder) Prepare(config interface{}) {
	tb.prepareCalled = true
	tb.prepareConfig = config
}

func (tb *TestBuilder) Run(b *Build, ui Ui) {
	tb.runCalled = true
	tb.runBuild = b
	tb.runUi = ui
}

func testBuild() *Build {
	return &Build{
		name: "test",
		builder: &TestBuilder{},
	}
}

func TestBuild_Prepare(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	build.Prepare(42)

	builder := build.builder.(*TestBuilder)

	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, 42, "prepare config should be 42")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := testUi()

	build := testBuild()
	build.Run(ui)

	builder := build.builder.(*TestBuilder)

	assert.True(builder.runCalled, "run should be called")
	assert.Equal(builder.runBuild, build, "run should be called with build")
	assert.Equal(builder.runUi, ui, "run should be called with ui")
}
