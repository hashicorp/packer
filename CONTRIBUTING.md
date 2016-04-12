# Contributing to Packer

**First:** if you're unsure or afraid of _anything_, just ask
or submit the issue or pull request anyways. You won't be yelled at for
giving your best effort. The worst that can happen is that you'll be
politely asked to change something. We appreciate any sort of contributions,
and don't want a wall of rules to get in the way of that.

However, for those individuals who want a bit more guidance on the
best way to contribute to the project, read on. This document will cover
what we're looking for. By addressing all the points we're looking for,
it raises the chances we can quickly merge or address your contributions.

## Issues

### Reporting an Issue

* Make sure you test against the latest released version. It is possible
  we already fixed the bug you're experiencing.

* Run the command with debug ouput with the environment variable
  `PACKER_LOG`. For example: `PACKER_LOG=1 packer build template.json`. Take
  the *entire* output and create a [gist](https://gist.github.com) for linking
  to in your issue. Packer should strip sensitive keys from the output,
  but take a look through just in case.

* Provide a reproducible test case. If a contributor can't reproduce an
  issue, then it dramatically lowers the chances it'll get fixed. And in
  some cases, the issue will eventually be closed.

* Respond promptly to any questions made by the Packer team to your issue.
  Stale issues will be closed.

### Issue Lifecycle

1. The issue is reported.

2. The issue is verified and categorized by a Packer collaborator.
   Categorization is done via tags. For example, bugs are marked as "bugs"
   and easy fixes are marked as "easy".

3. Unless it is critical, the issue is left for a period of time (sometimes
   many weeks), giving outside contributors a chance to address the issue.

4. The issue is addressed in a pull request or commit. The issue will be
   referenced in the commit message so that the code that fixes it is clearly
   linked.

5. The issue is closed.

## Setting up Go to work on Packer

If you have never worked with Go before, you will have to complete the
following steps in order to be able to compile and test Packer. These instructions target POSIX-like environments (Mac OS X, Linux, Cygwin, etc.) so you may need to adjust them for Windows or other shells.

1. [Download](https://golang.org/dl) and install Go. The instructions below
   are for go 1.6. Earlier versions of Go are no longer supported.

2. Set and export the `GOPATH` environment variable and update your `PATH`. For
   example, you can add to your `.bash_profile`.

    ```
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    ```

3. Download the Packer source (and its dependencies) by running `go get
   github.com/mitchellh/packer`. This will download the Packer source to
   `$GOPATH/src/github.com/mitchellh/packer`.

4. When working on packer `cd $GOPATH/src/github.com/mitchellh/packer` so you
   can run `make` and easily access other files.

5. Make your changes to the Packer source. You can run `make` in
   `$GOPATH/src/github.com/mitchellh/packer` to run tests and build the packer
   binary. Any compilation errors will be shown when the binaries are
   rebuilding. If you don't have `make` you can simply run `go build -o bin/packer .` from the project root.

6. After running building packer successfully, use
   `$GOPATH/src/github.com/mitchellh/packer/bin/packer` to build a machine and
   verify your changes work. For instance: `$GOPATH/src/github.com/mitchellh/packer/bin/packer build template.json`.

7. If everything works well and the tests pass, run `go fmt` on your code
   before submitting a pull-request.

### Tips for Working on Packer

#### Godeps

If you are submitting a change that requires a change in dependencies, DO NOT update the `vendor/` folder. This keeps the PR smaller and easier to review. Instead, please indicate which upstream has changed and which version we should be using. You _may_ do this using `Godeps/Godeps.json` but this is not required.

#### Running Unit Tests

You can run tests for individual packages using commands like this:

    $ make test TEST=./builder/amazon/...

#### Running Acceptance Tests

Packer has [acceptance tests](https://en.wikipedia.org/wiki/Acceptance_testing)
for various builders. These typically require an API key (AWS, GCE), or
additional software to be installed on your computer (VirtualBox, VMware).

If you're working on a feature of a builder or a new builder and want verify it
is functioning (and also hasn't broken anything else), we recommend running the
acceptance tests.

**Warning:** The acceptance tests create/destroy/modify *real resources*, which
may incur real costs in some cases. In the presence of a bug, it is technically
possible that broken backends could leave dangling data behind. Therefore,
please run the acceptance tests at your own risk. At the very least, we
recommend running them in their own private account for whatever builder you're
testing.

To run the acceptance tests, invoke `make testacc`:

    $ make testacc TEST=./builder/amazon/ebs
    ...

The `TEST` variable lets you narrow the scope of the acceptance tests to a
specific package / folder. The `TESTARGS` variable is recommended to filter
down to a specific resource to test, since testing all of them at once can
sometimes take a very long time.

Acceptance tests typically require other environment variables to be set for
things such as access keys. The test itself should error early and tell you
what to set, so it is not documented here.