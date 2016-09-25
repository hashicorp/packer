#!/usr/bin/env bats
#
# This tests the ansible provisioner on Google Cloud Provider (i.e.
# googlecompute). The teardown function will delete any images with the text
# "packerbats" within the name.

load test_helper
fixtures provisioner-ansible

# Required parameters
: ${GC_ACCOUNT_FILE:?}
: ${GC_PROJECT_ID:?}
command -v gcloud >/dev/null 2>&1 || {
    echo "'gcloud' must be installed" >&2
    exit 1
}

USER_VARS="${USER_VARS} -var account_file=${GC_ACCOUNT_FILE}"
USER_VARS="${USER_VARS} -var project_id=${GC_PROJECT_ID}"

# This tests if GCE has an image that contains the given parameter.
gc_has_image() {
    gcloud compute --format='table[no-heading](name)' --project=${GC_PROJECT_ID} images list \
        | grep $1 | wc -l
}

setup(){
    rm -f $FIXTURE_ROOT/ansible-test-id
    rm -f $FIXTURE_ROOT/ansible-server.key
    ssh-keygen -N "" -f $FIXTURE_ROOT/ansible-test-id
    ssh-keygen -N "" -f $FIXTURE_ROOT/ansible-server.key
}

teardown() {
    gcloud compute --format='table[no-heading](name)' --project=${GC_PROJECT_ID} images list \
        | grep packerbats \
        | xargs -n1 gcloud compute --project=${GC_PROJECT_ID} images delete

    rm -f $FIXTURE_ROOT/ansible-test-id
    rm -f $FIXTURE_ROOT/ansible-test-id.pub
    rm -f $FIXTURE_ROOT/ansible-server.key
    rm -f $FIXTURE_ROOT/ansible-server.key.pub
    rm -rf $FIXTURE_ROOT/fetched-dir
}

@test "ansible provisioner: build minimal.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-minimal")" -eq 1 ]
    diff -r dir fetched-dir/default/tmp/remote-dir > /dev/null
}

@test "ansible provisioner: build all_options.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/all_options.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-alloptions")" -eq 1 ]
    diff -r dir fetched-dir/packer-test/tmp/remote-dir > /dev/null
}

@test "ansible provisioner: build scp.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/scp.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-scp")" -eq 1 ]
    diff -r dir fetched-dir/default/tmp/remote-dir > /dev/null
}

@test "ansible provisioner: build scp-to-sftp.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/scp-to-sftp.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-scp-to-sftp")" -eq 1 ]
    diff -r dir fetched-dir/default/tmp/remote-dir > /dev/null
}

@test "ansible provisioner: build sftp.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/sftp.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-sftp")" -eq 1 ]
    diff -r dir fetched-dir/default/tmp/remote-dir > /dev/null
}

@test "ansible provisioner: build winrm.json" {
    cd $FIXTURE_ROOT
    run packer build ${USER_VARS} $FIXTURE_ROOT/winrm.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-winrm")" -eq 1 ]
    echo "packer does not support downloading files from download, skipping verification"
    #diff -r dir fetched-dir/default/tmp/remote-dir > /dev/null
}
