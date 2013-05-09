package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

type TestBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled bool
	runBuild Build
	runUi Ui
}

func (tb *TestBuilder) Prepare(config interface{}) error {
	tb.prepareCalled = true
	tb.prepareConfig = config
	return nil
}

func (tb *TestBuilder) Run(b Build, ui Ui) {
	tb.runCalled = true
	tb.runBuild = b
	tb.runUi = ui
}

func testBuild() Build {
	return &coreBuild{
		name: "test",
		builder: &TestBuilder{},
		rawConfig: 42,
	}
}

func testBuilder() *TestBuilder {
	return &TestBuilder{}
}

func TestBuild_Prepare(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	build := testBuild()
	builder := build.(*coreBuild).builder.(*TestBuilder)

	build.Prepare()
	assert.True(builder.prepareCalled, "prepare should be called")
	assert.Equal(builder.prepareConfig, 42, "prepare config should be 42")
}

func TestBuild_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := testUi()

	build := testBuild()
	build.Prepare()
	build.Run(ui)

	builder := build.(*coreBuild).builder.(*TestBuilder)

	assert.True(builder.runCalled, "run should be called")
	assert.Equal(builder.runBuild, build, "run should be called with build")
	assert.Equal(builder.runUi, ui, "run should be called with ui")
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
