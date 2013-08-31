package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

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

	hook := &MockHook{}

	mapping := make(map[string][]Hook)
	mapping["foo"] = []Hook{hook}
	dh := &DispatchHook{mapping}
	dh.Run("foo", nil, nil, 42)

	assert.True(hook.RunCalled, "run should be called")
	assert.Equal(hook.RunName, "foo", "should be proper event")
	assert.Equal(hook.RunData, 42, "should be correct data")
}
