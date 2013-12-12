#!/usr/bin/env bats
#
# This tests the amazon-ebs builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
fixtures provisioner-shell

setup() {
    cd $FIXTURE_ROOT
}

teardown() {
    aws_ami_cleanup
}

@test "shell provisioner: inline scripts" {
    run packer build $FIXTURE_ROOT/inline.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"HELLO I AM ubuntu"* ]]
    [[ "$output" == *"AND ANOTHER"* ]]
}

@test "shell provisioner: script" {
    run packer build $FIXTURE_ROOT/script.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"HELLO I AM DOG"* ]]
}

@test "shell provisioner: scripts" {
    run packer build $FIXTURE_ROOT/scripts.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"HELLO I AM DOG"* ]]
}
