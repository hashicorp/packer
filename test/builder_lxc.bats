#!/usr/bin/env bats
#
# This tests the lxc builder. The teardown function will
# delete any images in the output-lxc-* folders.

#load test_helper
#fixtures builder-lxc
FIXTURE_ROOT="$BATS_TEST_DIRNAME/fixtures/builder-lxc"

# Required parameters
command -v lxc-create >/dev/null 2>&1 || {
    echo "'lxc-create' must be installed via the lxc (or lxc1 for ubuntu >=16.04) package" >&2
    exit 1
}

teardown() {
    rm -rf output-lxc-*
}

@test "lxc: build centos minimal.json" {
    run packer build -var template_name=centos  $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ -f output-lxc-centos/rootfs.tar.gz ]
    [ -f output-lxc-centos/lxc-config ]
}


@test "lxc: build trusty minimal.json" {
    run packer build -var template_name=ubuntu -var template_parameters="SUITE=trusty" $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ -f output-lxc-ubuntu/rootfs.tar.gz ]
    [ -f output-lxc-ubuntu/lxc-config ]
}

@test "lxc: build debian minimal.json" {
    run packer build -var template_name=debian -var template_parameters="SUITE=jessie" $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ -f output-lxc-debian/rootfs.tar.gz ]
    [ -f output-lxc-debian/lxc-config ]
}
