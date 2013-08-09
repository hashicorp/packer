package command

import (
	"testing"
)

func TestBuildFiltersValidate(t *testing.T) {
	bf := new(BuildFilters)

	err := bf.Validate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Both set
	bf.Except = make([]string, 1)
	bf.Only = make([]string, 1)
	err = bf.Validate()
	if err == nil {
		t.Fatal("should error")
	}

	// One set
	bf.Except = make([]string, 1)
	bf.Only = make([]string, 0)
	err = bf.Validate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	bf.Except = make([]string, 0)
	bf.Only = make([]string, 1)
	err = bf.Validate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}
