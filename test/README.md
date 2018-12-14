# Packer Black-Box Tests

This folder contains tests that test Packer using a black-box approach:
`packer` is executed directly (with whatever is on the PATH) and certain
results are expected.

Tests are run using [Bats](https://github.com/sstephenson/bats), and therefore
Bash is required to run any tests.

**Warning:** Many of these tests run using AWS, and therefore have a
real-world cost associated with running the tests. Be aware of that prior
to running the tests. Additionally, many tests will leave left-over artifacts
(AMIs) that you'll have to manually clean up.

## Running Tests

### Required Software

Before running the tests, you'll need the following installed. If you're
running on macOS, most of these are available with `brew`:

* [Bats](https://github.com/sstephenson/bats)

* [AWS cli](http://aws.amazon.com/cli/) for AWS tests as well as most
  of the components.

* [gcutil](https://developers.google.com/compute/docs/gcutil/#install) for
  Google Compute Engine tests.

### Configuring Tests

**For tests that require AWS credentials:**

Set the following self-explanatory environmental variables:

* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

**For tests that test Google Compute Engine:**

Set the following environmental variables:

* `GC_BUCKET_NAME`
* `GC_ACCOUNT_FILE`
* `GC_PROJECT_ID`

### Running

These tests are meant to be run _one file at a time_. There are some
test files (such as the amazon-chroot builder test) that simply won't
run except in special environments, so running all test files will probably
never work.

If you're working on Packer and want to test that your change didn't
adversely affect something, try running only the test that is related to
your change.

```
$ bats builder_amazon_ebs.bats
```

Note: Working directory doesn't matter. You can call the bats test file
from any directory.
