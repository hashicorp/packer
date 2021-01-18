package command

import (
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

func TestFmt_Recursive(t *testing.T) {
	c := &FormatCommand{
		Meta: testMeta(t),
	}

	unformattedData, err := ioutil.ReadFile("test-fixtures/fmt/unformatted.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to open the unformatted fixture %s", err)
	}

	var subDir string
	subDir, err = ioutil.TempDir("test-fixtures/fmt", "sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(subDir)

	var superSubDir string
	superSubDir, err = ioutil.TempDir(subDir, "super_sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(superSubDir)

	tf, err := ioutil.TempFile(subDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create top level tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write(unformattedData)
	tf.Close()

	subTf, err := ioutil.TempFile(superSubDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create sub level tempfile for test %s", err)
	}
	defer os.Remove(subTf.Name())

	_, _ = subTf.Write(unformattedData)
	subTf.Close()

	args := []string{"-recursive=true", subDir}

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	formattedData, err := ioutil.ReadFile("test-fixtures/fmt/formatted.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to open the formatted fixture %s", err)
	}

	validateFileIsFormatted(t, formattedData, tf)
	validateFileIsFormatted(t, formattedData, subTf)

	//Testing with recursive flag off that sub directories are not formatted
	tf, err = ioutil.TempFile(subDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create top level tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write(unformattedData)
	tf.Close()

	subTf, err = ioutil.TempFile(superSubDir, "*.pkrvars.hcl")
	if err != nil {
		t.Fatalf("failed to create sub level tempfile for test %s", err)
	}
	defer os.Remove(subTf.Name())

	_, _ = subTf.Write(unformattedData)
	subTf.Close()

	args = []string{subDir}

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	validateFileIsFormatted(t, formattedData, tf)
	validateFileIsFormatted(t, unformattedData, subTf)
}

func validateFileIsFormatted(t *testing.T, formattedData []byte, testFile *os.File) {
	//lets re-read the tempfile which should now be formatted
	data, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	if diff := cmp.Diff(string(data), string(formattedData)); diff != "" {
		t.Errorf("Unexpected format tfData output %s", diff)
	}
}
