package command

import (
	"flag"
	"reflect"
	"testing"
)

func TestBuildOptionFlags(t *testing.T) {
	opts := new(BuildOptions)
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	BuildOptionFlags(fs, opts)

	args := []string{
		"-except=foo,bar,baz",
		"-only=a,b",
		"-var=foo=bar",
		"-var", "bar=baz",
		"-var=foo=bang",
		"-var-file=foo",
		"-var-file=bar",
	}

	err := fs.Parse(args)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(opts.Except, expected) {
		t.Fatalf("bad: %#v", opts.Except)
	}

	expected = []string{"a", "b"}
	if !reflect.DeepEqual(opts.Only, expected) {
		t.Fatalf("bad: %#v", opts.Only)
	}

	if len(opts.UserVars) != 2 {
		t.Fatalf("bad: %#v", opts.UserVars)
	}

	if opts.UserVars["foo"] != "bang" {
		t.Fatalf("bad: %#v", opts.UserVars)
	}

	if opts.UserVars["bar"] != "baz" {
		t.Fatalf("bad: %#v", opts.UserVars)
	}

	expected = []string{"foo", "bar"}
	if !reflect.DeepEqual(opts.UserVarFiles, expected) {
		t.Fatalf("bad: %#v", opts.UserVarFiles)
	}
}

func TestUserVarValue_implements(t *testing.T) {
	var raw interface{}
	raw = new(userVarValue)
	if _, ok := raw.(flag.Value); !ok {
		t.Fatalf("userVarValue should be a Value")
	}
}

func TestUserVarValueSet(t *testing.T) {
	sv := new(userVarValue)
	err := sv.Set("key=value")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	vars := map[string]string(*sv)
	if vars["key"] != "value" {
		t.Fatalf("Bad: %#v", vars)
	}

	// Empty value
	err = sv.Set("key=")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	vars = map[string]string(*sv)
	if vars["key"] != "" {
		t.Fatalf("Bad: %#v", vars)
	}

	// Equal in value
	err = sv.Set("key=foo=bar")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	vars = map[string]string(*sv)
	if vars["key"] != "foo=bar" {
		t.Fatalf("Bad: %#v", vars)
	}

	// No equal
	err = sv.Set("key")
	if err == nil {
		t.Fatal("should have error")
	}
}
