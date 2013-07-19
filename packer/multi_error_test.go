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

func TestMultiErrorAppend_MultiError(t *testing.T) {
	original := &MultiError{
		Errors: []error{errors.New("foo")},
	}

	result := MultiErrorAppend(original, errors.New("bar"))
	if len(result.Errors) != 2 {
		t.Fatalf("wrong len: %d", len(result.Errors))
	}
}

func TestMultiErrorAppend_NonMultiError(t *testing.T) {
	original := errors.New("foo")
	result := MultiErrorAppend(original, errors.New("bar"))
	if len(result.Errors) != 2 {
		t.Fatalf("wrong len: %d", len(result.Errors))
	}
}
