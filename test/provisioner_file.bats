#!/usr/bin/env bats
#
# This tests the amazon-ebs builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
fixtures provisioner-file

setup() {
    cd $FIXTURE_ROOT
}

teardown() {
    aws_ami_cleanup
}

@test "file provisioner: single file" {
    run packer build $FIXTURE_ROOT/file.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"24901 miles"* ]]
}

@test "file provisioner: directory (no trailing slash)" {
    run packer build $FIXTURE_ROOT/dir_no_trailing.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"337 miles"* ]]
}

@test "file provisioner: directory (with trailing slash)" {
    run packer build $FIXTURE_ROOT/dir_with_trailing.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"337 miles"* ]]
}

@test "file provisioner: single file through sftp" {
    run packer build $FIXTURE_ROOT/file_sftp.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"24901 miles"* ]]
}

@test "file provisioner: directory through sftp (no trailing slash)" {
    run packer build $FIXTURE_ROOT/dir_no_trailing_sftp.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"337 miles"* ]]
}

@test "file provisioner: directory through sftp (with trailing slash)" {
    run packer build $FIXTURE_ROOT/dir_with_trailing_sftp.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"337 miles"* ]]
}
