package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

func TestFmt(t *testing.T) {
	s := &strings.Builder{}
	ui := &packersdk.BasicUi{
		Writer: s,
	}
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	c.Ui = ui

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "formatted.pkr.hcl")}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	expected := ""
	assert.Equal(t, expected, strings.TrimSpace(s.String()))
}

func TestFmt_unformattedPKRVarsTemplate(t *testing.T) {
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "unformatted.pkrvars.hcl")}
	if code := c.Run(args); code != 3 {
		fatalCommand(t, c.Meta)
	}
}

func TestFmt_unfomattedTemlateDirectory(t *testing.T) {
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	args := []string{"-check=true", filepath.Join(testFixture("fmt"), "")}

	if code := c.Run(args); code != 3 {
		fatalCommand(t, c.Meta)
	}
}

type RecursiveTestCase struct {
	TestCaseName             string
	Recursion                bool
	TopLevelFilePreFormat    []byte
	LowerLevelFilePreFormat  []byte
	TopLevelFilePostFormat   []byte
	LowerLevelFilePostFormat []byte
}

func TestFmt_Recursive(t *testing.T) {
	unformattedData := []byte(`ami_filter_name ="amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners =[ "137112412989" ]

`)

	formattedData := []byte(`ami_filter_name   = "amzn2-ami-hvm-*-x86_64-gp2"
ami_filter_owners = ["137112412989"]

`)

	recursiveTestCases := []RecursiveTestCase{
		{
			TestCaseName:             "With Recursive flag on",
			Recursion:                true,
			TopLevelFilePreFormat:    unformattedData,
			LowerLevelFilePreFormat:  unformattedData,
			TopLevelFilePostFormat:   formattedData,
			LowerLevelFilePostFormat: formattedData,
		},
		{
			TestCaseName:             "With Recursive flag off",
			Recursion:                false,
			TopLevelFilePreFormat:    unformattedData,
			LowerLevelFilePreFormat:  unformattedData,
			TopLevelFilePostFormat:   formattedData,
			LowerLevelFilePostFormat: unformattedData,
		},
	}

	c := &FormatCommand{
		Meta: testMeta(t),
	}

	for _, tc := range recursiveTestCases {
		executeRecursiveTestCase(t, tc, c)
	}
}

func executeRecursiveTestCase(t *testing.T, tc RecursiveTestCase, c *FormatCommand) {
	// Creating temp directories and files
	subDir, err := ioutil.TempDir("test-fixtures/fmt", "sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(subDir)

	superSubDir, err := ioutil.TempDir(subDir, "super_sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(superSubDir)

	tf, err := ioutil.TempFile(subDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create top level tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write(tc.TopLevelFilePreFormat)
	tf.Close()

	data, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}
	fmt.Println(fmt.Sprintf("top level data: %v", data))

	subTf, err := ioutil.TempFile(superSubDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create sub level tempfile for test %s", err)
	}
	defer os.Remove(subTf.Name())

	_, _ = subTf.Write(tc.LowerLevelFilePreFormat)
	subTf.Close()

	var args []string
	if tc.Recursion {
		args = []string{"-recursive=true", subDir}
	} else {
		args = []string{subDir}
	}

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	validateFileIsFormatted(t, tc.TopLevelFilePostFormat, tf, tc)
	validateFileIsFormatted(t, tc.LowerLevelFilePostFormat, subTf, tc)
}

func validateFileIsFormatted(t *testing.T, formattedData []byte, testFile *os.File, tc RecursiveTestCase) {
	data, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	if diff := cmp.Diff(string(data), string(formattedData)); diff != "" {
		t.Errorf("Unexpected format tfData output on tc: %v, diff:  %s", tc.TestCaseName, diff)
	}
}
