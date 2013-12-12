#!/usr/bin/env bats
#
# This tests the amazon-ebs builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
fixtures amazon-ebs

teardown() {
    aws ec2 describe-images --owners self --output json --filters 'Name=tag:packer-test,Values=true' \
        | jq -r -M '.Images[]["ImageId"]' \
        | xargs -n1 aws ec2 deregister-image --image-id
}

@test "build minimal.json" {
    run packer build $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
}
