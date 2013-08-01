package common

import (
	"testing"
)

func TestCheckTemplates_Basic(t *testing.T) {
	t.Parallel()

	type S struct {
		A string
	}

	// Valid
	valid := S{
		A: "foo",
	}

	// Test valid case
	s := valid
	err := CheckTemplates(&s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test invalid
	s = valid
	s.A = "{{invalid}}"
	err = CheckTemplates(&s)
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestCheckTemplates_Map(t *testing.T) {
	type S struct {
		M map[string]string
	}

	s := &S{
		M: map[string]string{"valid": "valid"},
	}
	err := CheckTemplates(s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	s = &S{
		M: map[string]string{"{{invalid}}": "valid"},
	}
	err = CheckTemplates(s)
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestCheckTemplates_Nested(t *testing.T) {
	t.Parallel()

	// Test nested valid/invalid
	type S struct {
		A string
	}

	type S_nested struct {
		A S
	}

	sn := &S_nested{A: S{A: "foo"}}
	err := CheckTemplates(sn)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	sn = &S_nested{A: S{A: "{{invalid}}"}}
	err = CheckTemplates(sn)
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestCheckTemplates_Slice(t *testing.T) {
	t.Parallel()

	// Test slice valid/invalid
	type S_slice struct {
		A []string
		B int
	}

	ss := &S_slice{A: []string{"valid"}}
	err := CheckTemplates(ss)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	ss = &S_slice{A: []string{"{{invalid}}"}}
	err = CheckTemplates(ss)
	if err == nil {
		t.Fatal("error expected")
	}

	// Test slice of structs
	type S_sliceStruct struct {
		A []S_slice
	}

	st := &S_sliceStruct{
		A: []S_slice{
			S_slice{A: []string{"valid"}},
		},
	}
	err = CheckTemplates(st)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	st = &S_sliceStruct{
		A: []S_slice{
			S_slice{A: []string{"{{invalid}}"}},
		},
	}
	err = CheckTemplates(st)
	if err == nil {
		t.Fatal("error expected")
	}
}
