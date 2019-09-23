#!/usr/bin/env bats
#
# This tests the ssh communicator using AWS builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
verify_aws_cli
fixtures communicator-ssh

setup() {
    cd $FIXTURE_ROOT
}

teardown() {
    aws_ami_cleanup
}

@test "shell provisioner: local port tunneling" {
    run packer build $FIXTURE_ROOT/local-tunnel.json
    [ "$status" -eq 0 ]
    [[ "$output" == *"Connection to localhost port 10022 [tcp/*] succeeded"* ]]
}

@test "shell provisioner: remote port tunneling" {
    run packer build $FIXTURE_ROOT/remote-tunnel.json
    [ "$status" -eq 0 ]
    MY_LOCAL_IP=$(curl -s https://ifconfig.co/)
    [[ "$output" == *"$MY_LOCAL_IP"* ]]
}
