// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package acctest

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner/file"
	shellprovisioner "github.com/hashicorp/packer/provisioner/shell"
)

// TestEnvVar must be set to a non-empty value for acceptance tests to run.
const TestEnvVar = "PACKER_ACC"

// TestCase is a single set of tests to run for a backend. A TestCase
// should generally map 1:1 to each test method for your acceptance
// tests.
type TestCase struct {
	// Precheck, if non-nil, will be called once before the test case
	// runs at all. This can be used for some validation prior to the
	// test running.
	PreCheck func()

	// Builder is the Builder that will be tested. It will be available
	// as the "test" builder in the template.
	Builder packersdk.Builder

	// Template is the template contents to use.
	Template string

	// Check is called after this step is executed in order to test that
	// the step executed successfully. If this is not set, then the next
	// step will be called
	Check TestCheckFunc

	// Teardown will be called before the test case is over regardless
	// of if the test succeeded or failed. This should return an error
	// in the case that the test can't guarantee all resources were
	// properly cleaned up.
	Teardown TestTeardownFunc

	// If SkipArtifactTeardown is true, we will not attempt to destroy the
	// artifact created in this test run.
	SkipArtifactTeardown bool
	// If set, overrides the default provisioner store with custom provisioners.
	// This can be useful for running acceptance tests for a particular
	// provisioner using a specific builder.
	// Default provisioner store:
	// ProvisionerStore: packersdk.MapOfProvisioner{
	// 	"shell": func() (packersdk.Provisioner, error) { return &shellprovisioner.Provisioner{}, nil },
	// 	"file":  func() (packersdk.Provisioner, error) { return &file.Provisioner{}, nil },
	// },
	ProvisionerStore packersdk.MapOfProvisioner
}

// TestCheckFunc is the callback used for Check in TestStep.
type TestCheckFunc func([]packersdk.Artifact) error

// TestTeardownFunc is the callback used for Teardown in TestCase.
type TestTeardownFunc func() error

// TestT is the interface used to handle the test lifecycle of a test.
//
// Users should just use a *testing.T object, which implements this.
type TestT interface {
	Error(args ...interface{})
	Fatal(args ...interface{})
	Skip(args ...interface{})
}

type TestBuilderSet struct {
	packer.BuilderSet
	StartFn func(name string) (packersdk.Builder, error)
}

func (tbs TestBuilderSet) Start(name string) (packersdk.Builder, error) { return tbs.StartFn(name) }

// Test performs an acceptance test on a backend with the given test case.
//
// Tests are not run unless an environmental variable "PACKER_ACC" is
// set to some non-empty value. This is to avoid test cases surprising
// a user by creating real resources.
//
// Tests will fail unless the verbose flag (`go test -v`, or explicitly
// the "-test.v" flag) is set. Because some acceptance tests take quite
// long, we require the verbose flag so users are able to see progress
// output.
func Test(t TestT, c TestCase) {
	// We only run acceptance tests if an env var is set because they're
	// slow and generally require some outside configuration.
	if os.Getenv(TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			TestEnvVar))
		return
	}

	// We require verbose mode so that the user knows what is going on.
	if !testTesting && !testing.Verbose() {
		t.Fatal("Acceptance tests must be run with the -v flag on tests")
		return
	}

	// Run the PreCheck if we have it
	if c.PreCheck != nil {
		c.PreCheck()
	}

	// Parse the template
	log.Printf("[DEBUG] Parsing template...")
	tpl, err := template.Parse(strings.NewReader(c.Template))
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to parse template: %s", err))
		return
	}

	if c.ProvisionerStore == nil {
		c.ProvisionerStore = packersdk.MapOfProvisioner{
			"shell": func() (packersdk.Provisioner, error) { return &shellprovisioner.Provisioner{}, nil },
			"file":  func() (packersdk.Provisioner, error) { return &file.Provisioner{}, nil },
		}
	}
	// Build the core
	log.Printf("[DEBUG] Initializing core...")
	core := packer.NewCore(&packer.CoreConfig{
		Components: packer.ComponentFinder{
			PluginConfig: &packer.PluginConfig{
				Builders: TestBuilderSet{
					BuilderSet: packersdk.MapOfBuilder{
						"test": func() (packersdk.Builder, error) { return c.Builder, nil },
					},
					StartFn: func(n string) (packersdk.Builder, error) {
						if n == "test" {
							return c.Builder, nil
						}

						return nil, nil
					},
				},
				Provisioners: c.ProvisionerStore,
			},
		},
		Template: tpl,
	})
	diags := core.Initialize(packer.InitializeOptions{})
	if diags.HasErrors() {
		t.Fatal(fmt.Sprintf("Failed to init core: %s", err))
		return
	}

	// Get the build
	log.Printf("[DEBUG] Retrieving 'test' build")
	build, err := core.Build("test")
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to get 'test' build: %s", err))
		return
	}

	// Prepare it
	log.Printf("[DEBUG] Preparing 'test' build")
	warnings, err := build.Prepare()
	if err != nil {
		t.Fatal(fmt.Sprintf("Prepare error: %s", err))
		return
	}
	if len(warnings) > 0 {
		t.Fatal(fmt.Sprintf(
			"Prepare warnings:\n\n%s",
			strings.Join(warnings, "\n")))
		return
	}

	// Run it! We use a temporary directory for caching and discard
	// any UI output. We discard since it shows up in logs anyways.
	log.Printf("[DEBUG] Running 'test' build")
	ui := &packersdk.BasicUi{
		Reader:      os.Stdin,
		Writer:      io.Discard,
		ErrorWriter: io.Discard,
		PB:          &packersdk.NoopProgressTracker{},
	}
	artifacts, err := build.Run(context.Background(), ui)
	if err != nil {
		t.Fatal(fmt.Sprintf("Run error:\n\n%s", err))
		goto TEARDOWN
	}

	// Check function
	if c.Check != nil {
		log.Printf("[DEBUG] Running check function")
		if err := c.Check(artifacts); err != nil {
			t.Fatal(fmt.Sprintf("Check error:\n\n%s", err))
			goto TEARDOWN
		}
	}

TEARDOWN:
	if !c.SkipArtifactTeardown {
		// Delete all artifacts
		for _, a := range artifacts {
			if err := a.Destroy(); err != nil {
				t.Error(fmt.Sprintf(
					"!!! ERROR REMOVING ARTIFACT '%s': %s !!!",
					a.String(), err))
			}
		}
	}

	// Teardown
	if c.Teardown != nil {
		log.Printf("[DEBUG] Running teardown function")
		if err := c.Teardown(); err != nil {
			t.Fatal(fmt.Sprintf("Teardown failure:\n\n%s", err))
			return
		}
	}
}

// This is for unit tests of this package.
var testTesting = false
