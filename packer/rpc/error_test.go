package rpc

import (
	"cgl.tideland.biz/asserts"
	"errors"
	"testing"
)

func TestBasicError_ImplementsError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r error
	e := &BasicError{""}

	assert.Implementor(e, &r, "should be an error")
}

func TestBasicError_MatchesMessage(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	err := errors.New("foo")
	wrapped := NewBasicError(err)

	assert.Equal(wrapped.Error(), err.Error(), "should have the same error")
}
