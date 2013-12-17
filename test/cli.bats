#!/usr/bin/env bats
#
# This tests the basic CLI functionality of Packer. It makes no network
# requests and should be very fast.

load test_helper

@test "cli: packer should show help" {
    run packer
    [ "$status" -eq 1 ]
    [[ "$output" == *"usage: packer"* ]]
}

@test "cli: packer version" {
    run packer version
    [ "$status" -eq 0 ]
    [[ "$output" == *"Packer v"* ]]

    run packer -v
    [ "$status" -eq 0 ]
    [[ "$output" == *"Packer v"* ]]

    run packer --version
    [ "$status" -eq 0 ]
    [[ "$output" == *"Packer v"* ]]
}

@test "cli: packer version show help" {
    run packer version -h
    [ "$status" -eq 0 ]
    [[ "$output" == *"usage: packer version"* ]]

    run packer version --help
    [ "$status" -eq 0 ]
    [[ "$output" == *"usage: packer version"* ]]
}
