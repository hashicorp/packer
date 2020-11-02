package command

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/manifest"
	shell_local_pp "github.com/hashicorp/packer/post-processor/shell-local"
	filep "github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
	shell_local "github.com/hashicorp/packer/provisioner/shell-local"
)

var (
	spaghettiCarbonara = `spaghetti
carbonara
`
	lasagna = `lasagna
tomato
mozza
cooking...
`
	tiramisu = `whip_york
mascarpone
whipped_egg_white
dress
`
)

func TestBuild(t *testing.T) {
	tc := []struct {
		name         string
		args         []string
		expectedCode int
		fileCheck
	}{
		{
			name: "var-args: json - json varfile sets an apple env var",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "apple.json"),
				filepath.Join(testFixture("var-arg"), "fruit_builder.json"),
			},
			fileCheck: fileCheck{expected: []string{"apple.txt"}},
		},
		{
			name: "json - json varfile sets an apple env var, " +
				"override with banana cli var",
			args: []string{
				"-var", "fruit=banana",
				"-var-file=" + filepath.Join(testFixture("var-arg"), "apple.json"),
				filepath.Join(testFixture("var-arg"), "fruit_builder.json"),
			},
			fileCheck: fileCheck{expected: []string{"banana.txt"}},
		},
		{
			name: "var-args: json - arg sets a pear env var",
			args: []string{
				"-var=fruit=pear",
				filepath.Join(testFixture("var-arg"), "fruit_builder.json"),
			},
			fileCheck: fileCheck{expected: []string{"pear.txt"}},
		},

		{
			name: "var-args: json - inexistent var file errs",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "potato.json"),
				filepath.Join(testFixture("var-arg"), "fruit_builder.json"),
			},
			expectedCode: 1,
			fileCheck:    fileCheck{notExpected: []string{"potato.txt"}},
		},

		{
			name: "var-args: hcl - inexistent json var file errs",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "potato.json"),
				testFixture("var-arg"),
			},
			expectedCode: 1,
			fileCheck:    fileCheck{notExpected: []string{"potato.txt"}},
		},

		{
			name: "var-args: hcl - inexistent hcl var file errs",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "potato.hcl"),
				testFixture("var-arg"),
			},
			expectedCode: 1,
			fileCheck:    fileCheck{notExpected: []string{"potato.hcl"}},
		},

		{
			name: "var-args: hcl - auto varfile sets a chocolate env var",
			args: []string{
				testFixture("var-arg"),
			},
			fileCheck: fileCheck{expected: []string{"chocolate.txt"}},
		},

		{
			name: "var-args: hcl - hcl varfile sets a apple env var",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "apple.hcl"),
				testFixture("var-arg"),
			},
			fileCheck: fileCheck{expected: []string{"apple.txt"}},
		},

		{
			name: "var-args: hcl - json varfile sets a apple env var",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "apple.json"),
				testFixture("var-arg"),
			},
			fileCheck: fileCheck{expected: []string{"apple.txt"}},
		},

		{
			name: "var-args: hcl - arg sets a tomato env var",
			args: []string{
				"-var=fruit=tomato",
				testFixture("var-arg"),
			},
			fileCheck: fileCheck{expected: []string{"tomato.txt"}},
		},

		{
			name: "source name: HCL",
			args: []string{
				"-parallel-builds=1", // to ensure order is kept
				testFixture("build-name-and-type"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"manifest.json": `{
  "builds": [
    {
      "name": "test",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": "",
      "custom_data": null
    },
    {
      "name": "potato",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": "",
      "custom_data": null
    }
  ],
  "last_run_uuid": ""
}`,
				},
			},
		},

		{
			name: "build name: JSON except potato",
			args: []string{
				"-except=potato",
				"-parallel-builds=1", // to ensure order is kept
				filepath.Join(testFixture("build-name-and-type"), "all.json"),
			},
			fileCheck: fileCheck{
				expected: []string{
					"null.test.txt",
					"null.potato.txt",
				},
				expectedContent: map[string]string{
					"manifest.json": `{
  "builds": [
    {
      "name": "test",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": "",
      "custom_data": null
    }
  ],
  "last_run_uuid": ""
}`,
				},
			},
		},

		{
			name: "build name: JSON only potato",
			args: []string{
				"-only=potato",
				"-parallel-builds=1", // to ensure order is kept
				filepath.Join(testFixture("build-name-and-type"), "all.json"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"manifest.json": `{
  "builds": [
    {
      "name": "potato",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": "",
      "custom_data": null
    }
  ],
  "last_run_uuid": ""
}`,
				},
			},
		},

		// only / except HCL2
		{
			name: "hcl - 'except' a build block",
			args: []string{
				"-except=my_build.*",
				testFixture("hcl-only-except"),
			},
			fileCheck: fileCheck{
				expected:    []string{"cherry.txt"},
				notExpected: []string{"chocolate.txt", "vanilla.txt"},
			},
		},

		{
			name: "hcl - 'only' a build block",
			args: []string{
				"-only=my_build.*",
				testFixture("hcl-only-except"),
			},
			fileCheck: fileCheck{
				notExpected: []string{"cherry.txt"},
				expected:    []string{"chocolate.txt", "vanilla.txt"},
			},
		},

		// recipes
		{
			name: "hcl - recipes",
			args: []string{
				testFixture("hcl", "recipes"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"NULL.spaghetti_carbonara.txt": spaghettiCarbonara,
					"NULL.lasagna.txt":             lasagna,
					"NULL.tiramisu.txt":            tiramisu,
				},
			},
		},

		{
			name: "hcl - recipes - except carbonara",
			args: []string{
				"-except", "recipes.null.spaghetti_carbonara",
				testFixture("hcl", "recipes"),
			},
			fileCheck: fileCheck{
				notExpected: []string{"NULL.spaghetti_carbonara.txt"},
				expectedContent: map[string]string{
					"NULL.lasagna.txt":  lasagna,
					"NULL.tiramisu.txt": tiramisu,
				},
			},
		},

		{
			name: "hcl - recipes - only lasagna",
			args: []string{
				"-only", "*lasagna",
				testFixture("hcl", "recipes"),
			},
			fileCheck: fileCheck{
				notExpected: []string{
					"NULL.spaghetti_carbonara.txt",
					"NULL.tiramisu.txt",
				},
				expectedContent: map[string]string{
					"NULL.lasagna.txt": lasagna,
				},
			},
		},
		{
			name: "hcl - recipes - only recipes",
			args: []string{
				"-only", "recipes.*",
				testFixture("hcl", "recipes"),
			},
			fileCheck: fileCheck{
				notExpected: []string{
					"NULL.tiramisu.txt",
				},
				expectedContent: map[string]string{
					"NULL.spaghetti_carbonara.txt": spaghettiCarbonara,
					"NULL.lasagna.txt":             lasagna,
				},
			},
		},
		{
			name: "hcl - build.name accessible",
			args: []string{
				filepath.Join(testFixture("build-name-and-type"), "buildname.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expected: []string{
					"pineapple.pizza.txt",
				},
			},
		},

		{
			name: "hcl - valid validation rule for default value",
			args: []string{
				filepath.Join(testFixture("hcl", "validation", "map")),
			},
			expectedCode: 0,
		},

		{
			name: "hcl - valid setting from varfile",
			args: []string{
				"-var-file", filepath.Join(testFixture("hcl", "validation", "map", "valid_value.pkrvars.hcl")),
				filepath.Join(testFixture("hcl", "validation", "map")),
			},
			expectedCode: 0,
		},

		{
			name: "hcl - invalid setting from varfile",
			args: []string{
				"-var-file", filepath.Join(testFixture("hcl", "validation", "map", "invalid_value.pkrvars.hcl")),
				filepath.Join(testFixture("hcl", "validation", "map")),
			},
			expectedCode: 1,
		},

		{
			name: "hcl - valid cmd ( invalid varfile bypased )",
			args: []string{
				"-var-file", filepath.Join(testFixture("hcl", "validation", "map", "invalid_value.pkrvars.hcl")),
				"-var", `image_metadata={key = "new_value", something = { foo = "bar" }}`,
				filepath.Join(testFixture("hcl", "validation", "map")),
			},
			expectedCode: 0,
		},

		{
			name: "hcl - invalid cmd ( valid varfile bypased )",
			args: []string{
				"-var-file", filepath.Join(testFixture("hcl", "validation", "map", "valid_value.pkrvars.hcl")),
				"-var", `image_metadata={key = "?", something = { foo = "wrong" }}`,
				filepath.Join(testFixture("hcl", "validation", "map")),
			},
			expectedCode: 1,
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.cleanup(t)
			run(t, tt.args, tt.expectedCode)
			tt.fileCheck.verify(t)
		})
	}
}

func Test_build_output(t *testing.T) {

	tc := []struct {
		command     []string
		env         []string
		expected    []string
		notExpected []string
		runtime     string
	}{
		{[]string{"build", "--color=false", testFixture("hcl", "reprepare", "shell-local.pkr.hcl")},
			nil,
			[]string{"null.example: hello from the NULL builder packeruser", "Build 'null.example' finished after"},
			[]string{},
			"posix"},
		{[]string{"build", "--color=false", testFixture("hcl", "reprepare", "shell-local-windows.pkr.hcl")},
			nil,
			[]string{"null.example: hello from the NULL  builder packeruser", "Build 'null.example' finished after"},
			[]string{},
			"windows"},
		{[]string{"build", "--color=false", testFixture("hcl", "provisioner-override.pkr.hcl")},
			nil,
			[]string{"null.example1: yes overridden", "null.example2: not overridden"},
			[]string{"null.example2: yes overridden", "null.example1: not overridden"},
			"posix"},
		{[]string{"build", "--color=false", testFixture("provisioners", "provisioner-override.json")},
			nil,
			[]string{"example1: yes overridden", "example2: not overridden"},
			[]string{"example2: yes overridden", "example1: not overridden"},
			"posix"},
	}

	for _, tc := range tc {
		if (runtime.GOOS == "windows") != (tc.runtime == "windows") {
			continue
		}
		t.Run(fmt.Sprintf("packer %s", tc.command), func(t *testing.T) {
			p := helperCommand(t, tc.command...)
			p.Env = append(p.Env, tc.env...)
			bs, err := p.Output()
			if err != nil {
				t.Fatalf("%v: %s", err, bs)
			}
			for _, expected := range tc.expected {
				if !strings.Contains(string(bs), expected) {
					t.Fatalf("Should contain output %s.\nReceived: %s", tc.expected, string(bs))
				}
			}
			for _, notExpected := range tc.notExpected {
				if strings.Contains(string(bs), notExpected) {
					t.Fatalf("Should NOT contain output %s.\nReceived: %s", tc.expected, string(bs))
				}
			}
		})
	}
}

func TestBuildOnlyFileCommaFlags(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel-builds=1",
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
	if code := c.Run([]string{"-parallel-builds=1", "-"}); code != 0 {
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
		"-parallel-builds=1",
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

func TestBuildProvisionAndPosProcessWithBuildVariablesSharing(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		filepath.Join(testFixture("build-variable-sharing"), "template.json"),
	}

	files := []string{
		"provisioner.Null.txt",
		"post-processor.Null.txt",
	}

	defer cleanup(files...)

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range files {
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
		"-parallel-builds=1",
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
	tc := []struct {
		name                     string
		args                     []string
		expectedFiles            []string
		buildNotExpectedFiles    []string
		postProcNotExpectedFiles []string
	}{
		{
			name: "JSON: except build and post-processor",
			args: []string{
				"-parallel-builds=1",
				"-except=chocolate,vanilla,tomato",
				filepath.Join(testFixture("build-only"), "template.json"),
			},
			expectedFiles:            []string{"apple.txt", "cherry.txt", "peach.txt"},
			buildNotExpectedFiles:    []string{"chocolate.txt", "vanilla.txt", "tomato.txt", "unnamed.txt"},
			postProcNotExpectedFiles: []string{"pear.txt, banana.txt"},
		},
		{
			name: "HCL2: except build and post-processor",
			args: []string{
				"-parallel-builds=1",
				"-except=file.chocolate,file.vanilla,tomato",
				filepath.Join(testFixture("build-only"), "template.pkr.hcl"),
			},
			expectedFiles:            []string{"apple.txt", "cherry.txt", "peach.txt"},
			buildNotExpectedFiles:    []string{"chocolate.txt", "vanilla.txt", "tomato.txt", "unnamed.txt"},
			postProcNotExpectedFiles: []string{"pear.txt, banana.txt"},
		},
		{
			name: "HCL2-JSON: except build and post-processor",
			args: []string{
				"-parallel-builds=1",
				"-except=file.chocolate,file.vanilla,tomato",
				filepath.Join(testFixture("build-only"), "template.pkr.json"),
			},
			expectedFiles:            []string{"apple.txt", "cherry.txt", "peach.txt"},
			buildNotExpectedFiles:    []string{"chocolate.txt", "vanilla.txt", "tomato.txt", "unnamed.txt"},
			postProcNotExpectedFiles: []string{"pear.txt, banana.txt"},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			defer cleanup()

			if code := c.Run(tt.args); code != 0 {
				fatalCommand(t, c.Meta)
			}

			for _, f := range tt.buildNotExpectedFiles {
				if fileExists(f) {
					t.Errorf("build not skipped: Expected NOT to find %s", f)
				}
			}
			for _, f := range tt.postProcNotExpectedFiles {
				if fileExists(f) {
					t.Errorf("post-processor not skipped: Expected NOT to find %s", f)
				}
			}
			for _, f := range tt.expectedFiles {
				if !fileExists(f) {
					t.Errorf("Expected to find %s", f)
				}
			}
		})
	}
}

func testHCLOnlyExceptFlags(t *testing.T, args, present, notPresent []string) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	defer cleanup()

	finalArgs := []string{"-parallel-builds=1"}
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, testFixture("hcl-only-except"))

	if code := c.Run(finalArgs); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range notPresent {
		if fileExists(f) {
			t.Errorf("Expected NOT to find %s", f)
		}
	}
	for _, f := range present {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}

func TestBuildCommand_HCLOnlyExceptOptions(t *testing.T) {
	tests := []struct {
		args       []string
		present    []string
		notPresent []string
	}{
		{
			[]string{"-only=chocolate"},
			[]string{},
			[]string{"chocolate.txt", "vanilla.txt", "cherry.txt"},
		},
		{
			[]string{"-only=*chocolate*"},
			[]string{"chocolate.txt"},
			[]string{"vanilla.txt", "cherry.txt"},
		},
		{
			[]string{"-except=*chocolate*"},
			[]string{"vanilla.txt", "cherry.txt"},
			[]string{"chocolate.txt"},
		},
		{
			[]string{"-except=*ch*"},
			[]string{"vanilla.txt"},
			[]string{"chocolate.txt", "cherry.txt"},
		},
		{
			[]string{"-only=*chocolate*", "-only=*vanilla*"},
			[]string{"chocolate.txt", "vanilla.txt"},
			[]string{"cherry.txt"},
		},
		{
			[]string{"-except=*chocolate*", "-except=*vanilla*"},
			[]string{"cherry.txt"},
			[]string{"chocolate.txt", "vanilla.txt"},
		},
		{
			[]string{"-only=my_build.file.chocolate"},
			[]string{"chocolate.txt"},
			[]string{"vanilla.txt", "cherry.txt"},
		},
		{
			[]string{"-except=my_build.file.chocolate"},
			[]string{"vanilla.txt", "cherry.txt"},
			[]string{"chocolate.txt"},
		},
		{
			[]string{"-only=file.cherry"},
			[]string{"cherry.txt"},
			[]string{"vanilla.txt", "chocolate.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.args), func(t *testing.T) {
			testHCLOnlyExceptFlags(t, tt.args, tt.present, tt.notPresent)
		})
	}
}

func TestBuildWithNonExistingBuilder(t *testing.T) {
	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	args := []string{
		"-parallel-builds=1",
		`-except=`,
		filepath.Join(testFixture("build-only"), "not-found.json"),
	}

	defer cleanup()

	if code := c.Run(args); code != 1 {
		t.Errorf("Expected to find exit code 1, found %d", code)
	}
	if !fileExists("chocolate.txt") {
		t.Errorf("Expected to find chocolate.txt")
	}
	if fileExists("vanilla.txt") {
		t.Errorf("NOT expected to find vanilla.tx")
	}
}

func run(t *testing.T, args []string, expectedCode int) {
	t.Helper()

	c := &BuildCommand{
		Meta: testMetaFile(t),
	}

	if code := c.Run(args); code != expectedCode {
		fatalCommand(t, c.Meta)
	}
}

type fileCheck struct {
	expected, notExpected []string
	expectedContent       map[string]string
}

func (fc fileCheck) cleanup(t *testing.T) {
	for _, file := range fc.expectedFiles() {
		t.Logf("removing %v", file)
		if err := os.Remove(file); err != nil {
			t.Errorf("failed to remove file %s: %v", file, err)
		}
	}
}

func (fc fileCheck) expectedFiles() []string {
	expected := fc.expected
	for file := range fc.expectedContent {
		expected = append(expected, file)
	}
	return expected
}

func (fc fileCheck) verify(t *testing.T) {
	for _, f := range fc.expectedFiles() {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
	for _, f := range fc.notExpected {
		if fileExists(f) {
			t.Errorf("Expected to not find %s", f)
		}
	}
	for file, expectedContent := range fc.expectedContent {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			t.Fatalf("ioutil.ReadFile: %v", err)
		}
		if diff := cmp.Diff(expectedContent, string(content)); diff != "" {
			t.Errorf("content of %s differs: %s", file, diff)
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
		BuilderStore: packer.MapOfBuilder{
			"file": func() (packer.Builder, error) { return &file.Builder{}, nil },
			"null": func() (packer.Builder, error) { return &null.Builder{}, nil },
		},
		ProvisionerStore: packer.MapOfProvisioner{
			"shell-local": func() (packer.Provisioner, error) { return &shell_local.Provisioner{}, nil },
			"shell":       func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
			"file":        func() (packer.Provisioner, error) { return &filep.Provisioner{}, nil },
		},
		PostProcessorStore: packer.MapOfPostProcessor{
			"shell-local": func() (packer.PostProcessor, error) { return &shell_local_pp.PostProcessor{}, nil },
			"manifest":    func() (packer.PostProcessor, error) { return &manifest.PostProcessor{}, nil },
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

func cleanup(moreFiles ...string) {
	os.RemoveAll("chocolate.txt")
	os.RemoveAll("vanilla.txt")
	os.RemoveAll("cherry.txt")
	os.RemoveAll("apple.txt")
	os.RemoveAll("peach.txt")
	os.RemoveAll("banana.txt")
	os.RemoveAll("pear.txt")
	os.RemoveAll("tomato.txt")
	os.RemoveAll("unnamed.txt")
	os.RemoveAll("roses.txt")
	os.RemoveAll("fuchsias.txt")
	os.RemoveAll("lilas.txt")
	os.RemoveAll("campanules.txt")
	os.RemoveAll("ducky.txt")
	os.RemoveAll("banana.txt")
	for _, file := range moreFiles {
		os.RemoveAll(file)
	}
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
		wantCfg      *BuildArgs
		wantExitCode int
	}{
		{fields{defaultMeta},
			args{[]string{"file.json"}},
			&BuildArgs{
				MetaArgs:       MetaArgs{Path: "file.json"},
				ParallelBuilds: math.MaxInt64,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel-builds=10", "file.json"}},
			&BuildArgs{
				MetaArgs:       MetaArgs{Path: "file.json"},
				ParallelBuilds: 10,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel-builds=1", "file.json"}},
			&BuildArgs{
				MetaArgs:       MetaArgs{Path: "file.json"},
				ParallelBuilds: 1,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel-builds=5", "file.json"}},
			&BuildArgs{
				MetaArgs:       MetaArgs{Path: "file.json"},
				ParallelBuilds: 5,
				Color:          true,
			},
			0,
		},
		{fields{defaultMeta},
			args{[]string{"-parallel-builds=1", "-parallel-builds=5", "otherfile.json"}},
			&BuildArgs{
				MetaArgs:       MetaArgs{Path: "otherfile.json"},
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
