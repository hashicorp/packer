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

## Issues

### Reporting an Issue

* Make sure you test against the latest released version. It is possible we
  already fixed the bug you're experiencing.

* Run the command with debug output with the environment variable `PACKER_LOG`.
  For example: `PACKER_LOG=1 packer build template.json`. Take the _entire_
  output and create a [gist](https://gist.github.com) for linking to in your
  issue. Packer should strip sensitive keys from the output, but take a look
  through just in case.

* Provide a reproducible test case. If a contributor can't reproduce an issue,
  then it dramatically lowers the chances it'll get fixed. And in some cases,
  the issue will eventually be closed.

* Respond promptly to any questions made by the Packer team to your issue. Stale
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

5. The issue is closed.

## Setting up Go

If you have never worked with Go before, you will have to install its
runtime in order to build packer.

1. This project always releases from the latest version of golang. [Install go](https://golang.org/doc/install#install)

## Setting up Packer for dev

If/when you have go installed you can already `go get` packer and `make` in
order to compile and test Packer. These instructions target
POSIX-like environments (macOS, Linux, Cygwin, etc.) so you may need to
adjust them for Windows or other shells.
The instructions below are for go 1.7. or later.


1. Download the Packer source (and its dependencies) by running
   `go get github.com/hashicorp/packer`. This will download the Packer source to
   `$GOPATH/src/github.com/hashicorp/packer`.

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
   `$GOPATH/src/github.com/hashicorp/packer/bin/packer build template.json`.

5. If everything works well and the tests pass, run `go fmt` on your code before
   submitting a pull-request.

### Opening an Pull Request

Thank you for contributing! When you are ready to open a pull-request, you will
need to [fork
Packer](https://github.com/hashicorp/packer#fork-destination-box), push your
changes to your fork, and then open a pull-request.

For example, my github username is `cbednarski`, so I would do the following:

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

#### Running Unit Tests

You can run tests for individual packages using commands like this:

```
make test TEST=./builder/amazon/...
```

#### Running Acceptance Tests

Packer has [acceptance tests](https://en.wikipedia.org/wiki/Acceptance_testing)
for various builders. These typically require an API key (AWS, GCE), or
additional software to be installed on your computer (VirtualBox, VMware).

If you're working on a new builder or builder feature and want to verify it is
functioning (and also hasn't broken anything else), we recommend running the
acceptance tests.

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
