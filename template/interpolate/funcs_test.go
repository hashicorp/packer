package interpolate

import (
	"os"
	"testing"
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
