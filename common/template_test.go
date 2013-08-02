package common

import (
	"testing"
)

func TestNewConfigTemplate(t *testing.T) {
	_, err := NewConfigTemplate(nil)
	if err == nil {
		t.Fatal("should err")
	}

	_, err = NewConfigTemplate(struct{}{})
	if err == nil {
		t.Fatal("should err")
	}

	_, err = NewConfigTemplate(new(int))
	if err == nil {
		t.Fatal("should err")
	}

	ct, err := NewConfigTemplate(&struct{}{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if ct == nil {
		t.Fatal("result should not be nil")
	}
}

func TestConfigTemplateCheck_Basic(t *testing.T) {
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
	ct, err := NewConfigTemplate(&s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
	if err != nil {
		t.Fatalf("err: %p", err)
	}

	// Test invalid
	s = valid
	s.A = "{{invalid}}"
	ct, err = NewConfigTemplate(&s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestConfigTemplateCheck_Map(t *testing.T) {
	type S struct {
		M map[string]string
	}

	s := &S{
		M: map[string]string{"valid": "valid"},
	}
	ct, err := NewConfigTemplate(s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ct.Check(); err != nil {
		t.Fatalf("err: %s", err)
	}

	s = &S{
		M: map[string]string{"{{invalid}}": "valid"},
	}
	ct, err = NewConfigTemplate(s)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestConfigTemplateCheck_Nested(t *testing.T) {
	t.Parallel()

	// Test nested valid/invalid
	type S struct {
		A string
	}

	type S_nested struct {
		A S
	}

	sn := &S_nested{A: S{A: "foo"}}
	ct, err := NewConfigTemplate(sn)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ct.Check(); err != nil {
		t.Fatalf("err: %s", err)
	}

	sn = &S_nested{A: S{A: "{{invalid}}"}}
	ct, err = NewConfigTemplate(sn)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestConfigTemplateCheck_Slice(t *testing.T) {
	t.Parallel()

	// Test slice valid/invalid
	type S_slice struct {
		A []string
		B int
	}

	ss := &S_slice{A: []string{"valid"}}
	ct, err := NewConfigTemplate(ss)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ct.Check(); err != nil {
		t.Fatalf("err: %s", err)
	}

	ss = &S_slice{A: []string{"{{invalid}}"}}
	ct, err = NewConfigTemplate(ss)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
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
	ct, err = NewConfigTemplate(st)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ct.Check(); err != nil {
		t.Fatalf("err: %s", err)
	}

	st = &S_sliceStruct{
		A: []S_slice{
			S_slice{A: []string{"{{invalid}}"}},
		},
	}
	ct, err = NewConfigTemplate(st)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = ct.Check()
	if err == nil {
		t.Fatal("error expected")
	}
}

func TestConfigTemplateProcess(t *testing.T) {
	type S struct {
		Foo string
		Bar []string
		Baz map[string]string
	}

	config := &S{
		Foo: "foo",
		Bar: []string{"foo"},
		Baz: map[string]string{
			"foo": "foo",
		},
	}

	ct, err := NewConfigTemplate(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if err := ct.Process(); err != nil {
		t.Fatalf("err: %s", err)
	}

	if config.Foo != "bar" {
		t.Fatalf("bad value: %s", config.Foo)
	}

	if config.Bar[0] != "bar" {
		t.Fatalf("bad value: %s", config.Bar[0])
	}

	if config.Baz["bar"] != "bar" {
		t.Fatalf("bad value: %s", config.Baz["bar"])
	}
}
