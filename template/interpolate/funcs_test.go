package interpolate

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/version"
)

func TestFuncBuildName(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{build_name}}`,
			"foo",
		},
	}

	ctx := &Context{BuildName: "foo"}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncBuildType(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{build_type}}`,
			"foo",
		},
	}

	ctx := &Context{BuildType: "foo"}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncEnv(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{env "PACKER_TEST_ENV"}}`,
			`foo`,
		},

		{
			`{{env "PACKER_TEST_ENV_NOPE"}}`,
			``,
		},
	}

	os.Setenv("PACKER_TEST_ENV", "foo")
	defer os.Setenv("PACKER_TEST_ENV", "")

	ctx := &Context{EnableEnv: true}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncEnv_disable(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
		Error  bool
	}{
		{
			`{{env "PACKER_TEST_ENV"}}`,
			"",
			true,
		},
	}

	ctx := &Context{EnableEnv: false}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if (err != nil) != tc.Error {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncIsotime(t *testing.T) {
	ctx := &Context{}
	i := &I{Value: "{{isotime}}"}
	result, err := i.Render(ctx)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	val, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	currentTime := time.Now().UTC()
	if currentTime.Sub(val) > 2*time.Second {
		t.Fatalf("val: %v (current: %v)", val, currentTime)
	}
}

func TestFuncPwd(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{pwd}}`,
			wd,
		},
	}

	ctx := &Context{}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncTemplatePath(t *testing.T) {
	path := "foo/bar"
	expected, _ := filepath.Abs(filepath.Dir(path))

	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{template_dir}}`,
			expected,
		},
	}

	ctx := &Context{
		TemplatePath: path,
	}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncSplit(t *testing.T) {
	cases := []struct {
		Input         string
		Output        string
		ErrorExpected bool
	}{
		{
			`{{split build_name "-" 0}}`,
			"foo",
			false,
		},
		{
			`{{split build_name "-" 1}}`,
			"bar",
			false,
		},
		{
			`{{split build_name "-" 2}}`,
			"",
			true,
		},
	}

	ctx := &Context{BuildName: "foo-bar"}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if (err == nil) == tc.ErrorExpected {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncTimestamp(t *testing.T) {
	expected := strconv.FormatInt(InitTime.Unix(), 10)

	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{timestamp}}`,
			expected,
		},
	}

	ctx := &Context{}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncUser(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			`{{user "foo"}}`,
			`foo`,
		},

		{
			`{{user "what"}}`,
			``,
		},
	}

	ctx := &Context{
		UserVariables: map[string]string{
			"foo": "foo",
		},
	}
	for _, tc := range cases {
		i := &I{Value: tc.Input}
		result, err := i.Render(ctx)
		if err != nil {
			t.Fatalf("Input: %s\n\nerr: %s", tc.Input, err)
		}

		if result != tc.Output {
			t.Fatalf("Input: %s\n\nGot: %s", tc.Input, result)
		}
	}
}

func TestFuncPackerVersion(t *testing.T) {
	template := `{{packer_version}}`

	ctx := &Context{}
	i := &I{Value: template}

	result, err := i.Render(ctx)
	if err != nil {
		t.Fatalf("Input: %s\n\nerr: %s", template, err)
	}

	// Only match the X.Y.Z portion of the whole version string.
	if !strings.HasPrefix(result, version.Version) {
		t.Fatalf("Expected input to include: %s\n\nGot: %s",
			version.Version, result)
	}
}
