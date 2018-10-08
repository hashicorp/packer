package packer

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	configHelper "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template"
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
			"build-names-basic.hcl",
			nil,
			[]string{"something"},
		},
		{
			"build-names-func.json",
			nil,
			[]string{"TUBES"},
		},
		{
			"build-names-func.hcl",
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
	for _, f := range []string{"build-basic.json", "build-basic.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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
}

func TestCoreBuild_basicInterpolated(t *testing.T) {
	for _, f := range []string{"build-basic-interpolated.json", "build-basic-interpolated.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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
}

func TestCoreBuild_env(t *testing.T) {
	os.Setenv("PACKER_TEST_ENV", "test")
	defer os.Setenv("PACKER_TEST_ENV", "")

	for _, f := range []string{"build-env.json", "build-env.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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

		// Interpolate the config
		var result map[string]interface{}
		err = configHelper.Decode(&result, nil, b.PrepareConfig...)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if result["value"] != "test" {
			t.Fatalf("bad: %#v", result)
		}
	}
}

func TestCoreBuild_buildNameVar(t *testing.T) {
	for _, f := range []string{"build-var-build-name.json", "build-var-build-name.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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

		// Interpolate the config
		var result map[string]interface{}
		err = configHelper.Decode(&result, nil, b.PrepareConfig...)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if result["value"] != "test" {
			t.Fatalf("bad: %#v", result)
		}
	}
}

func TestCoreBuild_buildTypeVar(t *testing.T) {
	for _, f := range []string{"build-var-build-type.json", "build-var-build-type.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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

		// Interpolate the config
		var result map[string]interface{}
		err = configHelper.Decode(&result, nil, b.PrepareConfig...)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if result["value"] != "test" {
			t.Fatalf("bad: %#v", result)
		}
	}
}

func TestCoreBuild_nonExist(t *testing.T) {
	for _, f := range []string{"build-basic.json", "build-basic.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
		TestBuilder(t, config, "test")
		core := TestCore(t, config)

		_, err := core.Build("nope")
		if err == nil {
			t.Fatal("should error")
		}
	}
}

func TestCoreBuild_prov(t *testing.T) {
	for _, f := range []string{"build-prov.json", "build-prov.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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
}

func TestCoreBuild_provSkip(t *testing.T) {
	for _, f := range []string{"build-prov-skip.json", "build-prov-skip.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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
}

func TestCoreBuild_provSkipInclude(t *testing.T) {
	for _, f := range []string{"build-prov-skip-include.json", "build-prov-skip-include.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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
}

func TestCoreBuild_provOverride(t *testing.T) {
	for _, f := range []string{"build-prov-override.json", "build-prov-override.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
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

		found := false
		for _, raw := range p.PrepConfigs {
			if m, ok := raw.([]map[string]interface{}); ok {
				for _, m := range m {
					if _, ok := m["foo"]; ok {
						found = true
						break
					}
				}
			}
		}
		if !found {
			t.Fatal("override not called")
		}
	}
}

func TestCoreBuild_postProcess(t *testing.T) {
	for _, f := range []string{"build-pp.json", "build-pp.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
		b := TestBuilder(t, config, "test")
		p := TestPostProcessor(t, config, "test")
		core := TestCore(t, config)
		ui := TestUi(t)

		b.ArtifactId = "hello"
		p.ArtifactId = "goodbye"

		build, err := core.Build("test")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if _, err := build.Prepare(); err != nil {
			t.Fatalf("err: %s", err)
		}

		artifact, err := build.Run(ui, nil)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if len(artifact) != 1 {
			t.Fatalf("bad: %#v", artifact)
		}

		if artifact[0].Id() != p.ArtifactId {
			t.Fatalf("bad: %s", artifact[0].Id())
		}
		if p.PostProcessArtifact.Id() != b.ArtifactId {
			t.Fatalf("bad: %s", p.PostProcessArtifact.Id())
		}
	}
}

func TestCoreBuild_templatePath(t *testing.T) {
	for _, f := range []string{"build-template-path.json", "build-template-path.hcl"} {
		config := TestCoreConfig(t)
		testCoreTemplate(t, config, fixtureDir(f))
		b := TestBuilder(t, config, "test")
		core := TestCore(t, config)

		expected, _ := filepath.Abs("./test-fixtures")

		build, err := core.Build("test")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if _, err := build.Prepare(); err != nil {
			t.Fatalf("err: %s", err)
		}

		// Interpolate the config
		var result map[string]interface{}
		err = configHelper.Decode(&result, nil, b.PrepareConfig...)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		if result["value"] != expected {
			t.Fatalf("bad: %#v", result)
		}
	}
}

func TestCore_pushInterpolate(t *testing.T) {
	cases := []struct {
		File   string
		Vars   map[string]string
		Result template.Push
	}{
		{
			"push-vars.json",
			map[string]string{"foo": "bar"},
			template.Push{Name: "bar"},
		},
		{
			"push-vars.hcl",
			map[string]string{"foo": "bar"},
			template.Push{Name: "bar"},
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

		expected := core.Template.Push
		if !reflect.DeepEqual(expected, tc.Result) {
			t.Fatalf("err: %s\n\n%#v", tc.File, expected)
		}
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
		{
			"validate-dup-builder.hcl",
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
			"validate-req-variable.hcl",
			nil,
			true,
		},

		{
			"validate-req-variable.json",
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"validate-req-variable.hcl",
			map[string]string{"foo": "bar"},
			false,
		},

		// Min version good
		{
			"validate-min-version.json",
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"validate-min-version.hcl",
			map[string]string{"foo": "bar"},
			false,
		},

		{
			"validate-min-version-high.json",
			map[string]string{"foo": "bar"},
			true,
		},
		{
			"validate-min-version-high.hcl",
			map[string]string{"foo": "bar"},
			true,
		},
	}

	for i, tc := range cases {
		f, err := os.Open(fixtureDir(tc.File))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		tpl, err := template.Parse(f)
		f.Close()
		if err != nil {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}

		_, err = NewCore(&CoreConfig{
			Template:  tpl,
			Variables: tc.Vars,
			Version:   "1.0.0",
		})

		if (err != nil) != tc.Err {
			t.Errorf("[%d]err: %v\n\n%v", i, tc.File, err)
		}
	}
}

func TestSensitiveVars(t *testing.T) {
	cases := []struct {
		File          string
		Vars          map[string]string
		SensitiveVars []string
		Expected      string
		Err           bool
	}{
		// hardcoded
		{
			"sensitive-variables.json",
			map[string]string{"foo": "bar"},
			[]string{"foo"},
			"bar",
			false,
		},
		{
			"sensitive-variables.hcl",
			map[string]string{"foo": "bar"},
			[]string{"foo"},
			"bar",
			false,
		},
		// interpolated
		{
			"sensitive-variables.json",
			map[string]string{"foo": "{{build_name}}"},
			[]string{"foo"},
			"test",
			false,
		},
		{
			"sensitive-variables.hcl",
			map[string]string{"foo": "{{build_name}}"},
			[]string{"foo"},
			"test",
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

		_, err = NewCore(&CoreConfig{
			Template:  tpl,
			Variables: tc.Vars,
			Version:   "1.0.0",
		})

		if (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
		filtered := LogSecretFilter.get()
		if filtered[0] != tc.Expected && len(filtered) != 1 {
			t.Fatalf("not filtering sensitive vars; filtered is %#v", filtered)
		}
	}
}

func testComponentFinder() *ComponentFinder {
	builderFactory := func(n string) (Builder, error) { return new(MockBuilder), nil }
	ppFactory := func(n string) (PostProcessor, error) { return new(MockPostProcessor), nil }
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
