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
following steps in order to be able to compile and test Packer.

1. Install Go. Make sure the Go version is at least Go 1.2. Packer will not work with anything less than
   Go 1.2. On a Mac, you can `brew install go` to install Go 1.2.

2. Set and export the `GOPATH` environment variable. For example, you can
   add `export GOPATH=$HOME/Documents/golang` to your `.bash_profile`.

3. Download the Packer source (and its dependencies) by running
   `go get github.com/mitchellh/packer`. This will download the Packer
   source to `$GOPATH/src/github.com/mitchellh/packer`.

4. Make your changes to the Packer source. You can run `make` from the main
   source directory to recompile all the binaries. Any compilation errors
   will be shown when the binaries are rebuilding.

5. Test your changes by running `make test` and then running
   `$GOPATH/src/github.com/mitchellh/packer/bin/packer` to build a machine.

6. If everything works well and the tests pass, run `go fmt` on your code
   before submitting a pull request.
