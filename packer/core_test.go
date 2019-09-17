package packer

import (
	"context"
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

	artifact, err := build.Run(context.Background(), nil)
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

	artifact, err := build.Run(context.Background(), nil)
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

func TestCoreBuild_env(t *testing.T) {
	os.Setenv("PACKER_TEST_ENV", "test")
	defer os.Setenv("PACKER_TEST_ENV", "")

	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-env.json"))
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

func TestCoreBuild_buildNameVar(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-var-build-name.json"))
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

func TestCoreBuild_buildTypeVar(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-var-build-type.json"))
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

	artifact, err := build.Run(context.Background(), nil)
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

	artifact, err := build.Run(context.Background(), nil)
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

	artifact, err := build.Run(context.Background(), nil)
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

func TestCoreBuild_provOverride(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-prov-override.json"))
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

	artifact, err := build.Run(context.Background(), nil)
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
		if m, ok := raw.(map[string]interface{}); ok {
			if _, ok := m["foo"]; ok {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("override not called")
	}
}

func TestCoreBuild_postProcess(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-pp.json"))
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

	artifact, err := build.Run(context.Background(), ui)
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

func TestCoreBuild_templatePath(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-template-path.json"))
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

		// Min version good
		{
			"validate-min-version.json",
			map[string]string{"foo": "bar"},
			false,
		},

		{
			"validate-min-version-high.json",
			map[string]string{"foo": "bar"},
			true,
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
	}
}

// Tests that we can properly interpolate user variables defined within the
// packer template
func TestCore_InterpolateUserVars(t *testing.T) {
	cases := []struct {
		File     string
		Expected map[string]string
		Err      bool
	}{
		{
			"build-variables-interpolate.json",
			map[string]string{
				"foo":  "bar",
				"bar":  "bar",
				"baz":  "barbaz",
				"bang": "bangbarbaz",
			},
			false,
		},
		{
			"build-variables-interpolate2.json",
			map[string]string{},
			true,
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

		ccf, err := NewCore(&CoreConfig{
			Template: tpl,
			Version:  "1.0.0",
		})

		if (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
		if !tc.Err {
			for k, v := range ccf.variables {
				if tc.Expected[k] != v {
					t.Fatalf("Expected %s but got %s", tc.Expected[k], v)
				}
			}
		}
	}
}

// Tests that we can properly interpolate user variables defined within a
// var-file provided alongside the Packer template
func TestCore_InterpolateUserVars_VarFile(t *testing.T) {
	cases := []struct {
		File      string
		Variables map[string]string
		Expected  map[string]string
		Err       bool
	}{
		{
			// tests that we can interpolate from var files when var isn't set in
			// originating template
			"build-basic-interpolated.json",
			map[string]string{
				"name":   "gotta-{{user `my_var`}}",
				"my_var": "interpolate-em-all",
			},
			map[string]string{
				"name":   "gotta-interpolate-em-all",
				"my_var": "interpolate-em-all"},
			false,
		},
		{
			// tests that we can interpolate from var files when var is set in
			// originating template as required
			"build-basic-interpolated-required.json",
			map[string]string{
				"name":   "gotta-{{user `my_var`}}",
				"my_var": "interpolate-em-all",
			},
			map[string]string{
				"name":   "gotta-interpolate-em-all",
				"my_var": "interpolate-em-all"},
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

		ccf, err := NewCore(&CoreConfig{
			Template:  tpl,
			Version:   "1.0.0",
			Variables: tc.Variables,
		})

		if (err != nil) != tc.Err {
			t.Fatalf("err: %s\n\n%s", tc.File, err)
		}
		if !tc.Err {
			for k, v := range ccf.variables {
				if tc.Expected[k] != v {
					t.Fatalf("Expected value %s for key %s but got %s",
						tc.Expected[k], k, v)
				}
			}
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
		// interpolated
		{
			"sensitive-variables.json",
			map[string]string{"foo": "bar",
				"bang": "{{ user `foo`}}"},
			[]string{"bang"},
			"bar",
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

		// clear filter so it doesn't break other tests
		LogSecretFilter.s = make(map[string]struct{})
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
