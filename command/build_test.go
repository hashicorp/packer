// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-uuid"
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
	one = "1\n"
	two = "2\n"
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
			name: "var-args: json - nonexistent var file errs",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "potato.json"),
				filepath.Join(testFixture("var-arg"), "fruit_builder.json"),
			},
			expectedCode: 1,
			fileCheck:    fileCheck{notExpected: []string{"potato.txt"}},
		},

		{
			name: "var-args: hcl - nonexistent json var file errs",
			args: []string{
				"-var-file=" + filepath.Join(testFixture("var-arg"), "potato.json"),
				testFixture("var-arg"),
			},
			expectedCode: 1,
			fileCheck:    fileCheck{notExpected: []string{"potato.txt"}},
		},

		{
			name: "var-args: hcl - nonexistent hcl var file errs",
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
		{
			name: "hcl - execute and use datasource",
			args: []string{
				testFixture("hcl", "datasource.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"chocolate.txt": "chocolate",
				},
			},
		},
		{
			name: "hcl - dynamic source blocks in a build block",
			args: []string{
				testFixture("hcl", "dynamic", "build.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"dummy.txt":       "layers/base/main/files",
					"postgres/13.txt": "layers/base/main/files\nlayers/base/init/files\nlayers/postgres/files",
				},
				expected: []string{"dummy-fooo.txt", "dummy-baar.txt", "postgres/13-fooo.txt", "postgres/13-baar.txt"},
			},
		},

		{
			name: "hcl - variables can be used in shared post-processor fields",
			args: []string{
				testFixture("hcl", "var-in-pp-name.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"example1.1.txt": one,
					"example2.2.txt": two,
				},
				notExpected: []string{
					"example1.2.txt",
					"example2.1.txt",
				},
			},
		},
		{
			name: "hcl - using build variables in post-processor",
			args: []string{
				testFixture("hcl", "build-var-in-pp.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"example.2.txt": two,
				},
			},
		},

		{
			name: "hcl - test crash #11381",
			args: []string{
				testFixture("hcl", "nil-component-crash.pkr.hcl"),
			},
			expectedCode: 1,
		},
		{
			name: "hcl - using variables in build block",
			args: []string{
				testFixture("hcl", "vars-in-build-block.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"example.2.txt": two,
				},
			},
		},
		{
			name: "hcl - recursive local using input var",
			args: []string{
				testFixture("hcl", "recursive_local_with_input"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"hey.txt": "hello",
				},
			},
		},
		{
			name: "hcl - recursive local using an unset input var",
			args: []string{
				testFixture("hcl", "recursive_local_with_unset_input"),
			},
			fileCheck:    fileCheck{},
			expectedCode: 1,
		},
		{
			name: "hcl - var with default value empty object/list can be set",
			args: []string{
				testFixture("hcl", "empty_object"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo.txt": "yo",
				},
			},
		},
		{
			name: "hcl - unknown ",
			args: []string{
				testFixture("hcl", "data-source-validation.pkr.hcl"),
			},
			fileCheck: fileCheck{
				expectedContent: map[string]string{
					"foo.txt": "foo",
				},
				expected: []string{
					"s3cr3t",
				},
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.cleanup(t)
			t.Logf("Running build on %s", tt.args)
			run(t, tt.args, tt.expectedCode)
			tt.fileCheck.verify(t, "")
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
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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

func testHCLOnlyExceptFlags(t *testing.T, args, present, notPresent []string, expectReturn int) {
	c := &BuildCommand{
		Meta: TestMetaFile(t),
	}

	defer cleanup()

	finalArgs := []string{"-parallel-builds=1"}
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, testFixture("hcl-only-except"))

	if code := c.Run(finalArgs); code != expectReturn {
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

func TestHCL2PostProcessorForceFlag(t *testing.T) {
	t.Helper()

	UUID, _ := uuid.GenerateUUID()
	// Manifest will only clean with force if the build's PACKER_RUN_UUID are different
	t.Setenv("PACKER_RUN_UUID", UUID)

	args := []string{
		filepath.Join(testFixture("hcl"), "force.pkr.hcl"),
	}
	fCheck := fileCheck{
		expectedContent: map[string]string{
			"manifest.json": fmt.Sprintf(`{
  "builds": [
    {
      "name": "potato",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": %q,
      "custom_data": null
    }
  ],
  "last_run_uuid": %q
}`, UUID, UUID),
		},
	}
	defer fCheck.cleanup(t)

	c := &BuildCommand{
		Meta: TestMetaFile(t),
	}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	fCheck.verify(t, "")

	// Second build should override previous manifest
	UUID, _ = uuid.GenerateUUID()
	t.Setenv("PACKER_RUN_UUID", UUID)

	args = []string{
		"-force",
		filepath.Join(testFixture("hcl"), "force.pkr.hcl"),
	}
	fCheck = fileCheck{
		expectedContent: map[string]string{
			"manifest.json": fmt.Sprintf(`{
  "builds": [
    {
      "name": "potato",
      "builder_type": "null",
      "files": null,
      "artifact_id": "Null",
      "packer_run_uuid": %q,
      "custom_data": null
    }
  ],
  "last_run_uuid": %q
}`, UUID, UUID),
		},
	}

	c = &BuildCommand{
		Meta: TestMetaFile(t),
	}
	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}
	fCheck.verify(t, "")
}

func TestBuildCommand_HCLOnlyExceptOptions(t *testing.T) {
	tests := []struct {
		args         []string
		present      []string
		notPresent   []string
		expectReturn int
	}{
		{
			[]string{"-only=chocolate"},
			[]string{},
			[]string{"chocolate.txt", "vanilla.txt", "cherry.txt"},
			1,
		},
		{
			[]string{"-only=*chocolate*"},
			[]string{"chocolate.txt"},
			[]string{"vanilla.txt", "cherry.txt"},
			0,
		},
		{
			[]string{"-except=*chocolate*"},
			[]string{"vanilla.txt", "cherry.txt"},
			[]string{"chocolate.txt"},
			0,
		},
		{
			[]string{"-except=*ch*"},
			[]string{"vanilla.txt"},
			[]string{"chocolate.txt", "cherry.txt"},
			0,
		},
		{
			[]string{"-only=*chocolate*", "-only=*vanilla*"},
			[]string{"chocolate.txt", "vanilla.txt"},
			[]string{"cherry.txt"},
			0,
		},
		{
			[]string{"-except=*chocolate*", "-except=*vanilla*"},
			[]string{"cherry.txt"},
			[]string{"chocolate.txt", "vanilla.txt"},
			0,
		},
		{
			[]string{"-only=my_build.file.chocolate"},
			[]string{"chocolate.txt"},
			[]string{"vanilla.txt", "cherry.txt"},
			0,
		},
		{
			[]string{"-except=my_build.file.chocolate"},
			[]string{"vanilla.txt", "cherry.txt"},
			[]string{"chocolate.txt"},
			0,
		},
		{
			[]string{"-only=file.cherry"},
			[]string{"cherry.txt"},
			[]string{"vanilla.txt", "chocolate.txt"},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s", tt.args), func(t *testing.T) {
			testHCLOnlyExceptFlags(t, tt.args, tt.present, tt.notPresent, tt.expectReturn)
		})
	}
}

func TestBuildWithNonExistingBuilder(t *testing.T) {
	c := &BuildCommand{
		Meta: TestMetaFile(t),
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
		Meta: TestMetaFile(t),
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

func (fc fileCheck) verify(t *testing.T, dir string) {
	for _, f := range fc.expectedFiles() {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("Expected to find %s: %v", f, err)
		}
	}
	for _, f := range fc.notExpected {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Errorf("Expected to not find %s", f)
		}
	}
	for file, expectedContent := range fc.expectedContent {
		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatalf("os.ReadFile: %v", err)
		}
		if diff := cmp.Diff(expectedContent, string(content)); diff != "" {
			t.Errorf("content of %s differs: %s", file, diff)
		}
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
	defaultMeta := TestMetaFile(t)
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

// TestProvisionerOnlyExcept checks that only/except blocks in provisioners/post-processors behave as expected
func TestProvisionerAndPostProcessorOnlyExcept(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		outputCheck  func(string, string) error
	}{
		{
			"json - only named build",
			[]string{
				"-only", "packer",
				testFixture("provisioners", "provisioner-only-except.json"),
			},
			0,
			func(out, _ string) error {
				if !strings.Contains(out, "packer provisioner packer and null") {
					return fmt.Errorf("missing expected provisioner output")
				}

				if !strings.Contains(out, "packer post-processor packer and null") {
					return fmt.Errorf("missing expected post-processor output")
				}

				if strings.Contains(out, "null post-processor") || strings.Contains(out, "null provisioner") {
					return fmt.Errorf("found traces of unnamed provisioner/post-processor, should not")
				}

				return nil
			},
		},
		{
			"json - only unnamed build",
			[]string{
				"-only", "null",
				testFixture("provisioners", "provisioner-only-except.json"),
			},
			0,
			func(out, _ string) error {
				if !strings.Contains(out, "null provisioner null and null") {
					return fmt.Errorf("missing expected provisioner output")
				}

				if !strings.Contains(out, "null post-processor null and null") {
					return fmt.Errorf("missing expected post-processor output")
				}

				if strings.Contains(out, "packer post-processor") || strings.Contains(out, "packer provisioner") {
					return fmt.Errorf("found traces of named provisioner/post-processor, should not")
				}

				return nil
			},
		},
		{
			"hcl - only one source build",
			[]string{
				"-only", "null.packer",
				testFixture("provisioners", "provisioner-only-except.pkr.hcl"),
			},
			0,
			func(out, _ string) error {
				if !strings.Contains(out, "packer provisioner packer and null") {
					return fmt.Errorf("missing expected provisioner output")
				}

				if !strings.Contains(out, "packer post-processor packer and null") {
					return fmt.Errorf("missing expected post-processor output")
				}

				if strings.Contains(out, "other post-processor") || strings.Contains(out, "other provisioner") {
					return fmt.Errorf("found traces of other provisioner/post-processor, should not")
				}

				return nil
			},
		},
		{
			"hcl - only other build",
			[]string{
				"-only", "null.other",
				testFixture("provisioners", "provisioner-only-except.pkr.hcl"),
			},
			0,
			func(out, _ string) error {
				if !strings.Contains(out, "other provisioner other and null") {
					return fmt.Errorf("missing expected provisioner output")
				}

				if !strings.Contains(out, "other post-processor other and null") {
					return fmt.Errorf("missing expected post-processor output")
				}

				if strings.Contains(out, "packer post-processor") || strings.Contains(out, "packer provisioner") {
					return fmt.Errorf("found traces of \"packer\" source provisioner/post-processor, should not")
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BuildCommand{
				Meta: TestMetaFile(t),
			}

			exitCode := c.Run(tt.args)
			if exitCode != tt.expectedCode {
				t.Errorf("process exit code mismatch: expected %d, got %d",
					tt.expectedCode,
					exitCode)
			}

			out, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			err := tt.outputCheck(out, stderr)
			if err != nil {
				if len(out) != 0 {
					t.Logf("command stdout: %q", out)
				}

				if len(stderr) != 0 {
					t.Logf("command stderr: %q", stderr)
				}
				t.Error(err.Error())
			}
		})
	}
}

// TestBuildCmd aims to test the build command, with output validation
func TestBuildCmd(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		outputCheck  func(string, string) error
	}{
		{
			name: "hcl - no build block error",
			args: []string{
				testFixture("hcl", "no_build.pkr.hcl"),
			},
			expectedCode: 1,
			outputCheck: func(_, err string) error {
				if !strings.Contains(err, "Error: Missing build block") {
					return fmt.Errorf("expected 'Error: Missing build block' in output, did not find it")
				}

				nbErrs := strings.Count(err, "Error: ")
				if nbErrs != 1 {
					return fmt.Errorf(
						"error: too many errors in stdout for build block, expected 1, got %d",
						nbErrs)
				}

				return nil
			},
		},
		{
			name: "hcl - undefined var set in pkrvars",
			args: []string{
				testFixture("hcl", "variables", "ref_non_existing"),
			},
			expectedCode: 0,
			outputCheck: func(out, err string) error {
				nbWarns := strings.Count(out, "Warning: ")
				if nbWarns != 0 {
					return fmt.Errorf(
						"error: too many warnings in build output, expected 0, got %d",
						nbWarns)
				}

				nbErrs := strings.Count(err, "Error: ")
				if nbErrs != 0 {
					return fmt.Errorf("error: expected build to succeed without errors, got %d",
						nbErrs)
				}
				return nil
			},
		},
		{
			name: "hcl - build block without source",
			args: []string{
				testFixture("hcl", "build_no_source.pkr.hcl"),
			},
			expectedCode: 1,
			outputCheck: func(_, err string) error {
				if !strings.Contains(err, "Error: missing source reference") {
					return fmt.Errorf("expected 'Error: missing source reference' in output, did not find it")
				}

				nbErrs := strings.Count(err, "Error: ")
				if nbErrs != 1 {
					return fmt.Errorf(
						"error: too many errors in stderr for build, expected 1, got %d",
						nbErrs)
				}

				logRegex := regexp.MustCompile("on.*build_no_source.pkr.hcl line 1")
				if !logRegex.MatchString(err) {
					return fmt.Errorf("error: missing context for error message")
				}

				return nil
			},
		},
		{
			name: "hcl - exclude post-processor, expect no warning",
			args: []string{
				"-except", "manifest",
				testFixture("hcl", "test_except_manifest.pkr.hcl"),
			},
			expectedCode: 0,
			outputCheck: func(out, err string) error {
				for _, stream := range []string{out, err} {
					if strings.Contains(stream, "Warning: an 'except' option was passed, but did not match any build") {
						return fmt.Errorf("Unexpected warning for build no match with except")
					}

					if strings.Contains(stream, "Running post-processor:") {
						return fmt.Errorf("Should not run post-processors, but found one")
					}
				}

				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &BuildCommand{
				Meta: TestMetaFile(t),
			}

			exitCode := c.Run(tt.args)
			if exitCode != tt.expectedCode {
				t.Errorf("process exit code mismatch: expected %d, got %d",
					tt.expectedCode,
					exitCode)
			}

			out, stderr := GetStdoutAndErrFromTestMeta(t, c.Meta)
			err := tt.outputCheck(out, stderr)
			if err != nil {
				if len(out) != 0 {
					t.Logf("command stdout: %q", out)
				}

				if len(stderr) != 0 {
					t.Logf("command stderr: %q", stderr)
				}
				t.Error(err.Error())
			}
		})
	}
}
