package packer

import (
	"context"
	"errors"
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

func TestCoreBuild_IgnoreTemplateVariables(t *testing.T) {
	os.Setenv("PACKER_TEST_ENV", "test")
	defer os.Setenv("PACKER_TEST_ENV", "")

	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-ignore-template-variable.json"))
	core := TestCore(t, config)

	if core.variables["http_ip"] != "{{ .HTTPIP }}" {
		t.Fatalf("bad: User variable http_ip={{ .HTTPIP }} should not be interpolated")
	}

	if core.variables["var"] != "test_{{ .PACKER_TEST_TEMP }}" {
		t.Fatalf("bad: User variable var should be half interpolated to var=test_{{ .PACKER_TEST_TEMP }} but was var=%s", core.variables["var"])
	}

	if core.variables["array_var"] != "us-west-1,us-west-2" {
		t.Fatalf("bad: User variable array_var should be \"us-west-1,us-west-2\" but was %s", core.variables["var"])
	}

	build, err := core.Build("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := build.Prepare(); err != nil {
		t.Fatalf("err: %s", err)
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
		{"validate-dup-builder.json", nil, true},

		// Required variable not set
		{"validate-req-variable.json", nil, true},
		{"validate-req-variable.json", map[string]string{"foo": "bar"}, false},

		// Min version good
		{"validate-min-version.json", map[string]string{"foo": "bar"}, false},
		{"validate-min-version-high.json", map[string]string{"foo": "bar"}, true},
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
			if tc.Err == false {
				t.Fatalf("Error interpolating %s: Expected no error, but got: %s", tc.File, err)
			} else {
				t.Fatalf("Error interpolating %s: Expected an error, but got: %s", tc.File, err)
			}

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

// Normally I wouldn't test a little helper function, but it's regex.
func TestIsDoneInterpolating(t *testing.T) {
	cases := []struct {
		inputString  string
		expectedBool bool
		expectedErr  bool
	}{
		// Many of these tests are just exercising the regex to make sure it
		// doesnt get confused by different kinds of whitespace
		{"charmander-{{ user `spacesaroundticks` }}", false, false},
		{"pidgey-{{ user `partyparrot`}}", false, false},
		{"jigglypuff-{{ user`notickspaaces`}}", false, false},
		{"eevee-{{user`nospaces`}}", false, false},
		{"staryu-{{  user  `somanyspaces`  }}", false, false},
		{"{{  user  `somanyspaces`  }}-{{isotime}}", false, false},
		// Make sure that we only flag on "user" when it's in the right set of
		// brackets, in a properly declared template engine format
		{"missingno-{{ user `missingbracket` }", true, false},
		{"missing2-{user ``missingopenbrackets }}", true, false},
		{"wat-userjustinname", true, false},
		// Any functions that aren't "user" should have already been properly
		// interpolated by the time this is called, so these cases aren't
		// realistic. That said, this makes it clear that this function doesn't
		// care about anything but the user function
		{"pokemon-{{ isotime }}", true, false},
		{"squirtle-{{ env `water`}}", true, false},
		{"bulbasaur-notinterpolated", true, false},
		{"extra-{{thisfunc `user`}}", true, false},
	}
	for _, tc := range cases {
		done, err := isDoneInterpolating(tc.inputString)
		if (err != nil) != tc.expectedErr {
			t.Fatalf("Test case failed. Error: %s expected error: "+
				"%t test string: %s", err, tc.expectedErr, tc.inputString)
		}
		if done != tc.expectedBool {
			t.Fatalf("Test case failed. inputString: %s. "+
				"Expected done = %t but got done = %t", tc.inputString,
				tc.expectedBool, done)
		}
	}
}

func TestEnvAndFileVars(t *testing.T) {
	os.Setenv("INTERPOLATE_TEST_ENV_1", "bulbasaur")
	os.Setenv("INTERPOLATE_TEST_ENV_3", "/path/to/nowhere")
	os.Setenv("INTERPOLATE_TEST_ENV_2", "5")
	os.Setenv("INTERPOLATE_TEST_ENV_4", "bananas")

	f, err := os.Open(fixtureDir("complex-recursed-env-user-var-file.json"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl, err := template.Parse(f)
	f.Close()
	if err != nil {
		t.Fatalf("err: %s\n\n%s", "complex-recursed-env-user-var-file.json", err)
	}

	ccf, err := NewCore(&CoreConfig{
		Template: tpl,
		Version:  "1.0.0",
		Variables: map[string]string{
			"var_1":     "partyparrot",
			"var_2":     "{{user `env_1`}}-{{user `env_2`}}{{user `env_3`}}-{{user `var_1`}}",
			"final_var": "{{user `env_1`}}/{{user `env_2`}}/{{user `env_4`}}{{user `env_3`}}-{{user `var_1`}}/vmware/{{user `var_2`}}.vmx",
		},
	})
	expected := map[string]string{
		"var_1":     "partyparrot",
		"var_2":     "bulbasaur-5/path/to/nowhere-partyparrot",
		"final_var": "bulbasaur/5/bananas/path/to/nowhere-partyparrot/vmware/bulbasaur-5/path/to/nowhere-partyparrot.vmx",
		"env_1":     "bulbasaur",
		"env_2":     "5",
		"env_3":     "/path/to/nowhere",
		"env_4":     "bananas",
	}
	if err != nil {
		t.Fatalf("err: %s\n\n%s", "complex-recursed-env-user-var-file.json", err)
	}
	for k, v := range ccf.variables {
		if expected[k] != v {
			t.Fatalf("Expected value %s for key %s but got %s",
				expected[k], k, v)
		}
	}

	// Clean up env vars
	os.Unsetenv("INTERPOLATE_TEST_ENV_1")
	os.Unsetenv("INTERPOLATE_TEST_ENV_3")
	os.Unsetenv("INTERPOLATE_TEST_ENV_2")
	os.Unsetenv("INTERPOLATE_TEST_ENV_4")
}

func testCoreTemplate(t *testing.T, c *CoreConfig, p string) {
	tpl, err := template.ParseFile(p)
	if err != nil {
		t.Fatalf("err: %s\n\n%s", p, err)
	}

	c.Template = tpl
}

func TestCoreBuild_provRetry(t *testing.T) {
	config := TestCoreConfig(t)
	testCoreTemplate(t, config, fixtureDir("build-prov-retry.json"))
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

	ui := testUi()
	p.ProvFunc = func(ctx context.Context) error {
		return errors.New("failed")
	}

	artifact, err := build.Run(context.Background(), ui)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if len(artifact) != 1 {
		t.Fatalf("bad: %#v", artifact)
	}

	if artifact[0].Id() != b.ArtifactId {
		t.Fatalf("bad: %s", artifact[0].Id())
	}
	if !p.ProvRetried {
		t.Fatal("provisioner should retry")
	}
}
