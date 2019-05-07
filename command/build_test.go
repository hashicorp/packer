package command

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/packer"
	shell_local "github.com/hashicorp/packer/post-processor/shell-local"
)

func TestBuildOnlyFileCommaFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel=false",
		"-only=chocolate,vanilla",
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"chocolate.txt", "vanilla.txt",
		"apple.txt", "peach.txt", "pear.txt", "unnamed.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}

	if fileExists("cherry.txt") {
		t.Error("Expected NOT to find cherry.txt")
	}

	if !fileExists("tomato.txt") {
		t.Error("Expected to find tomato.txt")
	}
}

func TestBuildStdin(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}
	f, err := os.Open(filepath.Join(testFixture("build-only"), "template.json"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stdin := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = stdin }()

	defer cleanup()
	if code := c.Run([]string{"-parallel=false", "-"}); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"vanilla.txt", "cherry.txt", "chocolate.txt",
		"unnamed.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}

func TestBuildOnlyFileMultipleFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel=false",
		"-only=chocolate",
		"-only=cherry",
		"-only=apple", // ignored
		"-only=peach", // ignored
		"-only=pear",  // ignored
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"vanilla.txt", "tomato.txt"} {
		if fileExists(f) {
			t.Errorf("Expected NOT to find %s", f)
		}
	}
	for _, f := range []string{"chocolate.txt", "cherry.txt",
		"apple.txt", "peach.txt", "pear.txt", "unnamed.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}

func TestBuildEverything(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel=false",
		`-except=`,
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"chocolate.txt", "vanilla.txt", "tomato.txt",
		"apple.txt", "cherry.txt", "pear.txt", "peach.txt", "unnamed.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}

func TestBuildExceptFileCommaFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel=false",
		"-except=chocolate,vanilla",
		filepath.Join(testFixture("build-only"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"chocolate.txt", "vanilla.txt", "tomato.txt",
		"unnamed.txt"} {
		if fileExists(f) {
			t.Errorf("Expected NOT to find %s", f)
		}
	}
	for _, f := range []string{"apple.txt", "cherry.txt", "pear.txt", "peach.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}

// fileExists returns true if the filename is found
func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		Builder: func(n string) (packer.Builder, error) {
			return &file.Builder{}, nil
		},
		PostProcessor: func(n string) (packer.PostProcessor, error) {
			return &shell_local.PostProcessor{}, nil
		},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaFile(t *testing.T) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: testCoreConfigBuilder(t),
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func cleanup() {
	os.RemoveAll("chocolate.txt")
	os.RemoveAll("vanilla.txt")
	os.RemoveAll("cherry.txt")
	os.RemoveAll("apple.txt")
	os.RemoveAll("peach.txt")
	os.RemoveAll("pear.txt")
	os.RemoveAll("tomato.txt")
	os.RemoveAll("unnamed.txt")
	os.RemoveAll("roses.txt")
	os.RemoveAll("fuchsias.txt")
	os.RemoveAll("lilas.txt")
	os.RemoveAll("campanules.txt")
}

func TestBuildCommand_ParseArgs(t *testing.T) {
	defaultMeta := testMetaFile(t)
	type fields struct {
		Meta Meta
	}
	type args struct {
		args []string
	}
	tests := []struct {
		fields       fields
		args         args
		wantCfg      Config
		wantExitCode int
	}{
		{fields{defaultMeta},
			args{[]string{"file.json"}},
			Config{
				Path:           "file.json",
				ParallelBuilds: math.MaxInt64,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel=true", "file.json"}},
			Config{
				Path:           "file.json",
				ParallelBuilds: math.MaxInt64,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel=false", "file.json"}},
			Config{
				Path:           "file.json",
				ParallelBuilds: 1,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel-builds=5", "file.json"}},
			Config{
				Path:           "file.json",
				ParallelBuilds: 5,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel=false", "-parallel-builds=5", "otherfile.json"}},
			Config{
				Path:           "otherfile.json",
				ParallelBuilds: 5,
				Color:          true,
			},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.args.args), func(t *testing.T) {
			c := &BuildCommand{
				Meta: tt.fields.Meta,
			}
			gotCfg, gotExitCode := c.ParseArgs(tt.args.args)
			if diff := cmp.Diff(gotCfg, tt.wantCfg); diff != "" {
				t.Fatalf("BuildCommand.ParseArgs() unexpected cfg %s", diff)
			}
			if gotExitCode != tt.wantExitCode {
				t.Fatalf("BuildCommand.ParseArgs() gotExitCode = %v, want %v", gotExitCode, tt.wantExitCode)
			}
		})
	}
}
