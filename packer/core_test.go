package packer

import (
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/packer/template"
)

func TestCoreBuildNames(t *testing.T) {
	cases := []struct {
		File   string
		Vars   map[string]string
		Result []string
	}{
		{
			"build-names-basic.json",
			nil,
			[]string{"something"},
		},

		{
			"build-names-func.json",
			nil,
			[]string{"TUBES"},
		},
	}

	for _, tc := range cases {
		tpl, err := template.ParseFile(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		core, err := NewCore(&CoreConfig{
			Template:  tpl,
			Variables: tc.Vars,
		})
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		names := core.BuildNames()
		if !reflect.DeepEqual(names, tc.Result) {
			t.Fatalf("err: %s\n\n%#v", tc.File, names)
		}
	}
}

func TestCoreBuild_basic(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-basic.json"))
	b := TestBuilder(t, config, "test")
	core := TestCore(t, config)

	b.ArtifactId = "hello"

	build, err := core.Build("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact, err := build.Run(nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
}

func TestCoreBuild_basicInterpolated(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-basic-interpolated.json"))
	b := TestBuilder(t, config, "test")
	core := TestCore(t, config)

	b.ArtifactId = "hello"

	build, err := core.Build("NAME")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact, err := build.Run(nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
}

func TestCoreBuild_nonExist(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-basic.json"))
	TestBuilder(t, config, "test")
	core := TestCore(t, config)

	_, err := core.Build("nope")
	if err == nil {
		t.Fatal("should error")
	}
}

func TestCoreBuild_prov(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-prov.json"))
	b := TestBuilder(t, config, "test")
	p := TestProvisioner(t, config, "test")
	core := TestCore(t, config)

	b.ArtifactId = "hello"

	build, err := core.Build("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact, err := build.Run(nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
	if !p.ProvCalled {
		t.Fatal("provisioner not called")
	}
}

func TestCoreBuild_provSkip(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-prov-skip.json"))
	b := TestBuilder(t, config, "test")
	p := TestProvisioner(t, config, "test")
	core := TestCore(t, config)

	b.ArtifactId = "hello"

	build, err := core.Build("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact, err := build.Run(nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
	if p.ProvCalled {
		t.Fatal("provisioner should not be called")
	}
}

func TestCoreBuild_provSkipInclude(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-prov-skip-include.json"))
	b := TestBuilder(t, config, "test")
	p := TestProvisioner(t, config, "test")
	core := TestCore(t, config)

	b.ArtifactId = "hello"

	build, err := core.Build("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
	}

	artifact, err := build.Run(nil, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
	if !p.ProvCalled {
		t.Fatal("provisioner should be called")
	}
}

func TestCoreValidate(t *testing.T) {
	cases := []struct {
		File string
		Vars map[string]string
		Err  bool
	}{
		{
			"validate-dup-builder.json",
			nil,
			true,
		},

		// Required variable not set
		{
			"validate-req-variable.json",
			nil,
			true,
		},

		{
			"validate-req-variable.json",
			map[string]string{"foo": "bar"},
			false,
		},
	}

	for _, tc := range cases {
		f, err := os.Open(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		tpl, err := template.Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		core, err := NewCore(&CoreConfig{
			Template:  tpl,
			Variables: tc.Vars,
		})
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		if err := core.Validate(); (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
	}
}

func testComponentFinder() *ComponentFinder {
	builderFactory := func(n string) (Builder, error) { return new(MockBuilder), nil }
	ppFactory := func(n string) (PostProcessor, error) { return new(TestPostProcessor), nil }
	provFactory := func(n string) (Provisioner, error) { return new(MockProvisioner), nil }
	return &ComponentFinder{
		Builder:       builderFactory,
		PostProcessor: ppFactory,
		Provisioner:   provFactory,
	}
}

func testCoreTemplate(t *testing.T, c *CoreConfig, p string) {
	tpl, err := template.ParseFile(p)
	if err != nil {
		t.Fatalf("err: %s\n\n%s", p, err)
	}

	c.Template = tpl
}
