# Contributing to Packer

**First:** if you're unsure or afraid of _anything_, just ask or submit the
issue or pull request anyway. You won't be yelled at for giving your best
effort. The worst that can happen is that you'll be politely asked to change
something. We appreciate any sort of contributions, and don't want a wall of
rules to get in the way of that.

However, for those individuals who want a bit more guidance on the best way to
contribute to the project, read on. This document will cover what we're looking
for. By addressing all the points we're looking for, it raises the chances we
can quickly merge or address your contributions.

When contributing in any way to the Packer project (new issue, PR, etc), please
be aware that our team identifies with many gender pronouns. Please remember to
use nonbinary pronouns (they/them) and gender neutral language ("Hello folks")
when addressing our team. For more reading on our code of conduct, please see the
[HashiCorp community guidelines](https://www.hashicorp.com/community-guidelines).

## Issues

### Reporting an Issue

- Make sure you test against the latest released version. It is possible we
  already fixed the bug you're experiencing.

- Run the command with debug output with the environment variable `PACKER_LOG`.
  For example: `PACKER_LOG=1 packer build template.pkr.hcl`. Take the _entire_
  output and create a [gist](https://gist.github.com) for linking to in your
  issue. Packer should strip sensitive keys from the output, but take a look
  through just in case.

- Provide a reproducible test case. If a contributor can't reproduce an issue,
  then it dramatically lowers the chances it'll get fixed. And in some cases,
  the issue will eventually be closed.

- Respond promptly to any questions made by the Packer team to your issue. Stale
  issues will be closed.

### Issue Lifecycle

1. The issue is reported.

2. The issue is verified and categorized by a Packer collaborator.
   Categorization is done via tags. For example, bugs are marked as "bugs" and
   simple fixes are marked as "good first issue".

3. Unless it is critical, the issue is left for a period of time (sometimes many
   weeks), giving outside contributors a chance to address the issue.

4. The issue is addressed in a pull request or commit. The issue will be
   referenced in the commit message so that the code that fixes it is clearly
   linked.

5. Sometimes, if you have a specialized environment or use case, the maintainers
   may ask for your help to test the patch. You are able to download an
   experimental binary of Packer containing the Pull Request's patch via from
   the Pull Request page on GitHub. You can do this by scrolling to the
   "checks" section on GitHub, and clicking "details" on the
   "store_artifacts" check. This will take you to Packer's Circle CI page for
   the build, and you will be able to click a tab named "Artifacts" which will
   contain zipped Packer binaries for each major OS architecture.

6. The issue is closed.

## Setting up Go

If you have never worked with Go before, you will have to install its
runtime in order to build packer.

1. This project always releases from the latest version of golang.
[Install go](https://golang.org/doc/install#install) To properly build from
source, you need to have golang >= v1.20

## Setting up Packer for dev

If/when you have go installed you can already clone packer and `make` in
order to compile and test Packer. These instructions target
POSIX-like environments (macOS, Linux, Cygwin, etc.) so you may need to
adjust them for Windows or other shells.


1. Create a directory in your GOPATH for the code `mkdir -p $(go env GOPATH)/src/github.com/hashicorp && cd $_`
and clone the packer repository from GitHub into your GOPATH `git clone https://github.com/hashicorp/packer.git`
then change into the packer directory `cd packer`

2. When working on Packer, first `cd $GOPATH/src/github.com/hashicorp/packer`
   so you can run `make` and easily access other files. Run `make help` to get
   information about make targets.

3. Make your changes to the Packer source. You can run `make` in
   `$GOPATH/src/github.com/hashicorp/packer` to run tests and build the Packer
   binary. Any compilation errors will be shown when the binaries are
   rebuilding. If you don't have `make` you can simply run
   `go build -o bin/packer .` from the project root.

4. After running building Packer successfully, use
   `$GOPATH/src/github.com/hashicorp/packer/bin/packer` to build a machine and
   verify your changes work. For instance:
   `$GOPATH/src/github.com/hashicorp/packer/bin/packer build template.pkr.hcl`.

5. If everything works well and the tests pass, run `go fmt` on your code before
   submitting a pull-request.

### Windows Systems

On windows systems you need at least the [MinGW Tools](http://www.mingw.org/), e.g. install via [choco](https://chocolatey.org/):

```
choco install mingw -y
```

This installs the GCC compiler, as well as a `mingw32-make` which can be used wherever
this documentation mentions `make`

when building using `go` you also need to mention the windows
executable extension

```
go build -o bin/packer.exe
```

### Opening a Pull Request

Thank you for contributing! When you are ready to open a pull-request, you will
need to [fork
Packer](https://github.com/hashicorp/packer#fork-destination-box), push your
changes to your fork, and then open a pull-request.

For example, my GitHub username is `cbednarski`, so I would do the following:

```
git checkout -b f-my-feature
# Develop a patch.
git push https://github.com/cbednarski/Packer f-my-feature
```

From there, open your fork in your browser to open a new pull-request.

**Note:** Go infers package names from their file paths. This means `go build`
will break if you `git clone` your fork instead of using `go get` on the main
Packer project.

**Note:** See '[Working with
forks](https://help.github.com/articles/working-with-forks/)' for a better way
to use `git push ...`.

### Pull Request Lifecycle

1. You are welcome to submit your pull request for commentary or review before
   it is fully completed. Please prefix the title of your pull request with
   "[WIP]" to indicate this. It's also a good idea to include specific questions
   or items you'd like feedback on.

2. Once you believe your pull request is ready to be merged, you can remove any
   "[WIP]" prefix from the title and a core team member will review.

3. One of Packer's core team members will look over your contribution and
   either merge, or provide comments letting you know if there is anything left
   to do. We do our best to provide feedback in a timely manner, but it may take
   some time for us to respond. We may also have questions that we need answered
   about the code, either because something doesn't make sense to us or because
   we want to understand your thought process.

4. If we have requested changes, you can either make those changes or, if you
   disagree with the suggested changes, we can have a conversation about our
   reasoning and agree on a path forward. This may be a multi-step process. Our
   view is that pull requests are a chance to collaborate, and we welcome
   conversations about how to do things better. It is the contributor's
   responsibility to address any changes requested. While reviewers are happy to
   give guidance, it is unsustainable for us to perform the coding work necessary
   to get a PR into a mergeable state.

5. Once all outstanding comments and checklist items have been addressed, your
   contribution will be merged! Merged PRs will be included in the next
   Packer release. The core team takes care of updating the
   [CHANGELOG.md](../CHANGELOG.md) as they merge.

6. In rare cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

### Tips for Working on Packer

#### Getting Your Pull Requests Merged Faster

It is much easier to review pull requests that are:

1. Well-documented: Try to explain in the pull request comments what your
   change does, why you have made the change, and provide instructions for how
   to produce the new behavior introduced in the pull request. If you can,
   provide screen captures or terminal output to show what the changes look
   like. This helps the reviewers understand and test the change.

2. Small: Try to only make one change per pull request. If you found two bugs
   and want to fix them both, that's _awesome_, but it's still best to submit
   the fixes as separate pull requests. This makes it much easier for reviewers
   to keep in their heads all of the implications of individual code changes,
   and that means the PR takes less effort and energy to merge. In general, the
   smaller the pull request, the sooner reviewers will be able to make time to
   review it.

3. Passing Tests: Based on how much time we have, we may not review pull
   requests which aren't passing our tests. (Look below for advice on how to
   run unit tests). If you need help figuring out why tests are failing, please
   feel free to ask, but while we're happy to give guidance it is generally
   your responsibility to make sure that tests are passing. If your pull request
   changes an interface or invalidates an assumption that causes a bunch of
   tests to fail, then you need to fix those tests before we can merge your PR.

If we request changes, try to make those changes in a timely manner. Otherwise,
PRs can go stale and be a lot more work for all of us to merge in the future.

Even with everyone making their best effort to be responsive, it can be
time-consuming to get a PR merged. It can be frustrating to deal with
the back-and-forth as we make sure that we understand the changes fully. Please
bear with us, and please know that we appreciate the time and energy you put
into the project.

### PR Checks

The following checks run when a PR is opened:

- Contributor License Agreement (CLA): If this is your first contribution to Packer you will be asked to sign the CLA.
- Tests: tests include unit tests, documentation checks, and code formatting checks, and all checks must pass before a PR can be merged.

#### Working on forks

The easiest way to work on a fork is to set it as a remote of the Packer
project. After following the steps in "Setting up Go to work on Packer":

1. Navigate to the code:

   `cd $GOPATH/src/github.com/hashicorp/packer`

2. Add the remote by running:

   `git remote add <name of remote> <github url of fork>`

   For example:

   `git remote add mwhooker https://github.com/mwhooker/packer.git`

3. Checkout a feature branch:

   `git checkout -b new-feature`

4. Make changes.
5. (Optional) Push your changes to the fork:

   `git push -u <name of remote> new-feature`

This way you can push to your fork to create a PR, but the code on disk still
lives in the spot where the go cli tools are expecting to find it.

#### Go modules & go vendor

If you are submitting a change that requires new or updated dependencies,
please include them in `go.mod`/`go.sum` and in the `vendor/` folder. This
helps everything get tested properly in CI.

Note that you will need to use [go
mod](https://github.com/golang/go/wiki/Modules) to do this. This step is
recommended but not required.

Use `go get <project>` to add dependencies to the project and `go mod vendor`
to make vendored copy of dependencies. See [go mod quick
start](https://github.com/golang/go/wiki/Modules#quick-start) for examples.

Please only apply the minimal vendor changes to get your PR to work. Packer
does not attempt to track the latest version for each dependency.

#### Code generation

Packer relies on `go generate` to generate a [peg parser for boot
commands](https://github.com/hashicorp/packer/blob/master/packer-plugin-sdk/bootcommand/boot_command.go),
[docs](https://github.com/hashicorp/packer/blob/master/website/pages/partials/builder/amazon/chroot/_Config-not-required.mdx)
and HCL2's bridging code. Packer's testing suite will run `make generate-check`
to check that all the generated files Packer needs are what they should be.
`make generate` re-generates all these file and can take a while depending on
your machine's performances. To make it faster it is recommended to run
localized code generation. Say you are working on the Amazon builder: running
`go generate ./builder/amazon/...` will do that for you. Make sure that the
latest code generation tool is installed by running `make install-gen-deps`.

#### Code linting

Packer relies on [golangci-lint](https://github.com/golangci/golangci-lint) for linting its Go code base, excluding any generated code created by `go generate`. Linting is executed on new files during Travis builds via `make ci`; the linting of existing code base is only executed when running `make lint`. Linting a large project like Packer is an iterative process so existing code base will have issues that are actively being fixed; pull-requests that fix existing linting issues are always welcomed :smile:.

The main configuration for golangci-lint is the `.golangci.yml` in the project root. See `golangci-lint --help` for a list of flags that can be used to override the default configuration.

Run golangci-lint on the entire Packer code base.

```
make lint
```

Run golangci-lint on a single pkg or directory; PKG_NAME expands to /builder/amazon/...

```
make lint PKG_NAME=builder/amazon
```

Note: linting on Travis uses the `--new-from-rev` flag to only lint new files added within a branch or pull-request. To run this check locally you can use the `ci-lint` make target. See [golangci-lint in CI](https://github.com/golangci/golangci-lint#faq) for more information.

```
make ci-lint
```

#### Running Unit Tests

You can run tests for individual packages using commands like this:

```
make test TEST=./builder/amazon/...
```

#### Running Builder Acceptance Tests

Packer has [acceptance tests](https://en.wikipedia.org/wiki/Acceptance_testing)
for various builders. These typically require an API key (AWS, GCE), or
additional software to be installed on your computer (VirtualBox, VMware).

If you're working on a new builder or builder feature and want to verify it is
functioning (and also hasn't broken anything else), we recommend creating or
running the acceptance tests.

**Warning:** The acceptance tests create/destroy/modify _real resources_, which
may incur costs for real money. In the presence of a bug, it is possible that
resources may be left behind, which can cost money even though you were not
using them. We recommend running tests in an account used only for that purpose
so it is easy to see if there are any dangling resources, and so production
resources are not accidentally destroyed or overwritten during testing.

To run the acceptance tests, invoke `make testacc`:

```
make testacc TEST=./builder/amazon/ebs
...
```

The `TEST` variable lets you narrow the scope of the acceptance tests to a
specific package / folder. The `TESTARGS` variable is recommended to filter down
to a specific resource to test, since testing all of them at once can sometimes
take a very long time.

To run only a specific test, use the `-run` argument:

```
make testacc TEST=./builder/amazon/ebs TESTARGS="-run TestBuilderAcc_forceDeleteSnapshot"
```

Acceptance tests typically require other environment variables to be set for
things such as API tokens and keys. Each test should error and tell you which
credentials are missing, so those are not documented here.

#### Running Provisioner Acceptance Tests

**Warning:** The acceptance tests create/destroy/modify _real resources_, which
may incur costs for real money. In the presence of a bug, it is possible that
resources may be left behind, which can cost money even though you were not
using them. We recommend running tests in an account used only for that purpose
so it is easy to see if there are any dangling resources, and so production
resources are not accidentally destroyed or overwritten during testing.
Also, these typically require an API key (AWS, GCE), or additional software
to be installed on your computer (VirtualBox, VMware).

To run the Provisioners Acceptance Tests you should use the
**ACC_TEST_BUILDERS** environment variable to tell the tests which builder the
test should be run against.

Examples of usage:

- Run the Shell provisioner acceptance tests against the Amazon EBS builder.
    ```
    ACC_TEST_BUILDERS=amazon-ebs go test ./provisioner/shell/... -v -timeout=1h
    ```
- Do the same but using the Makefile
    ```
    ACC_TEST_BUILDERS=amazon-ebs make provisioners-acctest TEST=./provisioner/shell
    ```
- Run all provisioner acceptance tests against the Amazon EBS builder.
    ```
    ACC_TEST_BUILDERS=amazon-ebs make provisioners-acctest  TEST=./...
    ```
- Run all provisioner acceptance tests against all builders whenever they are compatible.
    ```
    ACC_TEST_BUILDERS=all make provisioners-acctest  TEST=./...
    ```

The **ACC_TEST_BUILDERS** env variable accepts a list of builders separated by
commas. (e.g. `ACC_TEST_BUILDERS=amazon-ebs,virtualbox-iso`)


#### Writing Provisioner Acceptance Tests

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
  "github.com/hashicorp/packer-plugin-sdk/acctest/provisioneracc"
  "github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
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

#### Debugging Plugins

Each packer plugin runs in a separate process and communicates via RPC over a
socket therefore using a debugger will not work (be complicated at least).

But most of the Packer code is really simple and easy to follow with PACKER_LOG
turned on. If that doesn't work adding some extra debug print outs when you have
homed in on the problem is usually enough.
