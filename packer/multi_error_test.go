package packer

import (
	"cgl.tideland.biz/asserts"
	"errors"
	"testing"
)

func TestMultiError_Impl(t *testing.T) {
	var raw interface{}
	raw = &MultiError{}
	if _, ok := raw.(error); !ok {
		t.Fatal("MultiError must implement error")
	}
}

func TestMultiErrorError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := `2 error(s) occurred:

* foo
* bar`

	errors := []error{
		errors.New("foo"),
		errors.New("bar"),
	}

	multi := &MultiError{errors}
	assert.Equal(multi.Error(), expected, "should have proper error")
}
