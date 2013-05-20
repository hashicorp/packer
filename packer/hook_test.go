package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

type TestHook struct {
	runCalled bool
	runComm   Communicator
	runData   interface{}
	runName   string
	runUi     Ui
}

func (t *TestHook) Run(name string, ui Ui, comm Communicator, data interface{}) {
	t.runCalled = true
	t.runComm = comm
	t.runData = data
	t.runName = name
	t.runUi = ui
}

func TestDispatchHook_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r Hook
	c := &DispatchHook{nil}

	assert.Implementor(c, &r, "should be a Hook")
}

func TestDispatchHook_Run_NoHooks(t *testing.T) {
	// Just make sure nothing blows up
	dh := &DispatchHook{make(map[string][]Hook)}
	dh.Run("foo", nil, nil, nil)
}

func TestDispatchHook_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	hook := &TestHook{}

	mapping := make(map[string][]Hook)
	mapping["foo"] = []Hook{hook}
	dh := &DispatchHook{mapping}
	dh.Run("foo", nil, nil, 42)

	assert.True(hook.runCalled, "run should be called")
	assert.Equal(hook.runName, "foo", "should be proper event")
	assert.Equal(hook.runData, 42, "should be correct data")
}
