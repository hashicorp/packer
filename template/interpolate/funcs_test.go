package interpolate

import (
	"os"
	"strconv"
	"testing"
	"time"
)

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

	ctx := &Context{DisableEnv: true}
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
		t.Fatalf("val: %d (current: %d)", val, currentTime)
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
