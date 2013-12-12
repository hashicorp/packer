#!/usr/bin/env bats
#
# This tests the amazon-ebs builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
fixtures amazon-ebs

teardown() {
    aws_ami_cleanup
}

@test "amazon-ebs: build minimal.json" {
    run packer build $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
}
