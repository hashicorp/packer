# Packer Black-Box Tests

This folder contains tests that test Packer using a black-box approach:
`packer` is executed directly (with whatever is on the PATH) and certain
results are expected.

Tests are run using [Bats](https://github.com/sstephenson/bats), and therefore
Bash is required to run any tests.
