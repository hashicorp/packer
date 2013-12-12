#!/usr/bin/env bats
#
# This tests the amazon-ebs builder. The teardown function will automatically
# delete any AMIs with a tag of `packer-test` being equal to "true" so
# be sure any test cases set this.

load test_helper
fixtures amazon-ebs

# This counts how many AMIs were copied to another region
aws_ami_region_copy_count() {
    aws ec2 describe-images --region $1 --owners self --output text \
        --filters 'Name=tag:packer-id,Values=ami_region_copy' \
        --query "Images[*].ImageId" \
        | wc -l
}

teardown() {
    aws_ami_cleanup 'us-east-1'
    aws_ami_cleanup 'us-west-1'
    aws_ami_cleanup 'us-west-2'
}

@test "amazon-ebs: build minimal.json" {
    run packer build $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
}

# @unit-testable
@test "amazon-ebs: AMI region copy" {
    run packer build $FIXTURE_ROOT/ami_region_copy.json
    [ "$status" -eq 0 ]
    [ "$(aws_ami_region_copy_count 'us-east-1')" -eq "1" ]
    [ "$(aws_ami_region_copy_count 'us-west-1')" -eq "1" ]
    [ "$(aws_ami_region_copy_count 'us-west-2')" -eq "1" ]
}
