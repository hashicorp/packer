package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCDPrepare(t *testing.T) {
	type testCases struct {
		CDConfig        CDConfig
		ErrExpected     bool
		Reason          string
		ExpectedCDFiles []string
	}
	tcs := []testCases{
		{
			CDConfig:        CDConfig{},
			ErrExpected:     false,
			Reason:          "TestNilCD: nil CD array should not fail",
			ExpectedCDFiles: []string{},
		},
		{
			CDConfig:        CDConfig{CDFiles: make([]string, 0)},
			ErrExpected:     false,
			Reason:          "TestEmptyArrayCD: empty CD array should never fail",
			ExpectedCDFiles: []string{},
		},
		{
			CDConfig:        CDConfig{CDFiles: []string{"extra_iso_config.go"}},
			ErrExpected:     false,
			Reason:          "TestExistingCDFile: array with existing CD should not fail",
			ExpectedCDFiles: []string{"extra_iso_config.go"},
		},
		{
			CDConfig:        CDConfig{CDFiles: []string{"does_not_exist.foo"}},
			ErrExpected:     true,
			Reason:          "TestNonExistingCDFile: array with non existing CD should return errors",
			ExpectedCDFiles: []string{"does_not_exist.foo"},
		},
		{
			CDConfig:        CDConfig{CDFiles: []string{"extra_iso_config*"}},
			ErrExpected:     false,
			Reason:          "TestGlobbingCDFile: Glob should work",
			ExpectedCDFiles: []string{"extra_iso_config.go", "extra_iso_config_test.go"},
		},
	}
	for _, tc := range tcs {
		c := tc.CDConfig
		errs := c.Prepare(nil)
		if (len(errs) != 0) != tc.ErrExpected {
			t.Fatal(tc.Reason)
		}
		assert.Equal(t, c.CDFiles, tc.ExpectedCDFiles)
	}
}

func TestMultiErrorCDFiles(t *testing.T) {
	c := CDConfig{
		CDFiles: []string{"extra_iso_config.foo", "extra_iso_config.go",
			"extra_iso_config.bar", "extra_iso_config_test.go", "extra_iso_config.baz"},
	}

	errs := c.Prepare(nil)
	if len(errs) == 0 {
		t.Fatal("array with non existing CD should return errors")
	}

	expectedErrors := 3
	if count := len(errs); count != expectedErrors {
		t.Fatalf("array with %v non existing CD should return %v errors but it is returning %v", expectedErrors, expectedErrors, count)
	}
}
