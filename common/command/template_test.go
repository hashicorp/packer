package command

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testTemplate() (*packer.Template, *packer.ComponentFinder) {
	tplData := `{
	"builders": [
	{
		"type": "foo"
	},
	{
		"type": "bar"
	}
	]
}`

	tpl, err := packer.ParseTemplate([]byte(tplData), nil)
	if err != nil {
		panic(err)
	}

	cf := &packer.ComponentFinder{
		Builder: func(string) (packer.Builder, error) { return new(packer.MockBuilder), nil },
	}

	return tpl, cf
}

func TestBuildOptionsBuilds(t *testing.T) {
	opts := new(BuildOptions)
	bs, err := opts.Builds(testTemplate())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(bs) != 2 {
		t.Fatalf("bad: %d", len(bs))
	}
}

func TestBuildOptionsBuilds_except(t *testing.T) {
	opts := new(BuildOptions)
	opts.Except = []string{"foo"}

	bs, err := opts.Builds(testTemplate())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(bs) != 1 {
		t.Fatalf("bad: %d", len(bs))
	}

	if bs[0].Name() != "bar" {
		t.Fatalf("bad: %s", bs[0].Name())
	}
}

func TestBuildOptionsBuilds_only(t *testing.T) {
	opts := new(BuildOptions)
	opts.Only = []string{"foo"}

	bs, err := opts.Builds(testTemplate())
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if len(bs) != 1 {
		t.Fatalf("bad: %d", len(bs))
	}

	if bs[0].Name() != "foo" {
		t.Fatalf("bad: %s", bs[0].Name())
	}
}

func TestBuildOptionsBuilds_exceptNonExistent(t *testing.T) {
	opts := new(BuildOptions)
	opts.Except = []string{"i-dont-exist"}

	_, err := opts.Builds(testTemplate())
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestBuildOptionsBuilds_onlyNonExistent(t *testing.T) {
	opts := new(BuildOptions)
	opts.Only = []string{"i-dont-exist"}

	_, err := opts.Builds(testTemplate())
	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestBuildOptionsValidate(t *testing.T) {
	bf := new(BuildOptions)

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

func TestBuildOptionsValidate_userVarFiles(t *testing.T) {
	bf := new(BuildOptions)

	err := bf.Validate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Non-existent file
	bf.UserVarFiles = []string{"ireallyshouldntexistanywhere"}
	err = bf.Validate()
	if err == nil {
		t.Fatal("should error")
	}
}
