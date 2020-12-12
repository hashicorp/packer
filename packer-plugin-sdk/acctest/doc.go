/*
The acctest package provides an acceptance testing framework for testing
builders and provisioners.

Writing Provisioner Acceptance Tests

Packer has implemented a `ProvisionerTestCase` structure to help write
provisioner acceptance tests.

```go
type ProvisionerTestCase struct {
  // Check is called after this step is executed in order to test that
  // the step executed successfully. If this is not set, then the next
  // step will be called
  Check func(*exec.Cmd, string) error
  // IsCompatible checks whether a provisioner is able to run against a
  // given builder type and guest operating system, and returns a boolean.
  // if it returns true, the test combination is okay to run. If false, the
  // test combination is not okay to run.
  IsCompatible func(builderType string, BuilderGuestOS string) bool
  // Name is the name of the test case. Be simple but unique and descriptive.
  Name string
  // Setup, if non-nil, will be called once before the test case
  // runs. This can be used for some setup like setting environment
  // variables, or for validation prior to the
  // test running. For example, you can use this to make sure certain
  // binaries are installed, or text fixtures are in place.
  Setup func() error
  // Teardown will be called before the test case is over regardless
  // of if the test succeeded or failed. This should return an error
  // in the case that the test can't guarantee all resources were
  // properly cleaned up.
  Teardown builderT.TestTeardownFunc
  // Template is the provisioner template to use.
  // The provisioner template fragment must be a json-formatted string
  // containing the provisioner definition but no other portions of a packer
  // template. For
  // example:
  //
  // ```json
  // {
  //  "type": "shell-local",
  //  "inline", ["echo hello world"]
  // }
  //```
  //
  // is a valid entry for "template" here, but the complete Packer template:
  //
  // ```json
  // {
  //  "provisioners": [
  //    {
  //      "type": "shell-local",
  //      "inline", ["echo hello world"]
  //    }
  //  ]
  // }
  // ```
  //
  // is invalid as input.
  //
  // You may provide multiple provisioners in the same template. For example:
  // ```json
  // {
  //  "type": "shell-local",
  //  "inline", ["echo hello world"]
  // },
  // {
  //  "type": "shell-local",
  //  "inline", ["echo hello world 2"]
  // }
  // ```
  Template string
  // Type is the type of provisioner.
  Type string
}

```

To start writing a new provisioner acceptance test, you should add a test file
named `provisioner_acc_test.go` in the same folder as your provisioner is
defined. Create a test case by implementing the above struct, and run it
by calling `provisioneracc.TestProvisionersAgainstBuilders(testCase, t)`

The following example has been adapted from a shell-local provisioner test:

```
import (
  "github.com/hashicorp/packer/packer-plugin-sdk/acctest/provisioneracc"
  "github.com/hashicorp/packer/packer-plugin-sdk/acctest/testutils"
)

// ...

func TestAccShellProvisioner_basic(t *testing.T) {
  // Create a json template fragment containing just the provisioners you want
  // to run.
  templateString := `{
    "type": "shell-local",
    "script": "test-fixtures/script.sh",
    "max_retries" : 5
}`

  // instantiate a test case.
  testCase := &provisioneracc.ProvisionerTestCase{
    IsCompatible: func() bool {return true},
    Name:         "shell-local-provisioner-basic",
    Teardown: func() error {
      testutils.CleanupFiles("test-fixtures/file.txt")
      return nil
    },
    Template: templateString,
    Type:     "shell-local",
    Check: func(buildcommand *exec.Cmd, logfile string) error {
      if buildcommand.ProcessState != nil {
        if buildcommand.ProcessState.ExitCode() != 0 {
          return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
        }
      }
      filecontents, err := loadFile("file.txt")
      if err != nil {
        return err
      }
      if !strings.Contains(filecontents, "hello") {
        return fmt.Errorf("file contents were wrong: %s", filecontents)
      }
      return nil
    },
  }

  provisioneracc.TestProvisionersAgainstBuilders(testCase, t)
}

```


After writing the struct and implementing the interface, now is time to write the test that will run all
of this code you wrote. Your test should be like:

```go
func TestShellProvisioner(t *testing.T) {
	acc.TestProvisionersPreCheck("shell", t)
	acc.TestProvisionersAgainstBuilders(new(ShellProvisionerAccTest), t)
}
```

The method `TestProvisionersAgainstBuilders` will run the provisioner against
all available and compatible builders. If there are not builders compatible with
the test you want to run, you can add a builder using the following steps:

Create a subdirectory in provisioneracc/test-fixtures for the type of builder
you are adding. In this subdirectory, add one json file containing a single
builder fragment. For example, one of our amazon-ebs builders is defined in
provisioneracc/test-fixtures/amazon-ebs/amazon-ebs.txt and contains:

```json
{
  "type": "amazon-ebs",
  "ami_name": "packer-acc-test",
  "instance_type": "t2.micro",
  "region": "us-east-1",
  "ssh_username": "ubuntu",
  "source_ami_filter": {
    "filters": {
      "virtualization-type": "hvm",
      "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
      "root-device-type": "ebs"
    },
    "owners": ["099720109477"],
    "most_recent": true
  },
  "force_deregister" : true,
  "tags": {
    "packer-test": "true"
  }
}
```

note that this fragment does not contain anything other than a single builder
definition. The testing framework will combine this with the provisioner
fragment to create a working json template.

In order to tell the testing framework how to use this builder fragment, you
need to implement a `BuilderFixture` struct:

```go
type BuilderFixture struct {
  // Name is the name of the builder fixture.
  // Be simple and descriptive.
  Name string
  // Setup creates necessary extra test fixtures, and renders their values
  // into the BuilderFixture.Template.
  Setup func()
  // Template is the path to a builder template fragment.
  // The builder template fragment must be a json-formatted file containing
  // the builder definition but no other portions of a packer template. For
  // example:
  //
  // ```json
  // {
  //  "type": "null",
  //  "communicator", "none"
  // }
  //```
  //
  // is a valid entry for "template" here, but the complete Packer template:
  //
  // ```json
  // {
  //  "builders": [
  //    "type": "null",
  //    "communicator": "none"
  //  ]
  // }
  // ```
  //
  // is invalid as input.
  //
  // Only provide one builder template fragment per file.
  TemplatePath string

  // GuestOS says what guest os type the builder template fragment creates.
  // Valid values are "windows", "linux" or "darwin" guests.
  GuestOS string

  // HostOS says what host os type the builder is capable of running on.
  // Valid values are "any", windows", or "posix". If you set "posix", then
  // this builder can run on a "linux" or "darwin" platform. If you set
  // "any", then this builder can be used on any platform.
  HostOS string

  Teardown builderT.TestTeardownFunc
}
```
Implement this struct to the file "provisioneracc/builders.go", then add
the new implementation to the `BuildersAccTest` map in
`provisioneracc/provisioners.go`

Once you finish these steps, you should be ready to run your new provisioner
acceptance test by setting the name used in the BuildersAccTest map as your
`ACC_TEST_BUILDERS` environment variable.
*/

package acctest
