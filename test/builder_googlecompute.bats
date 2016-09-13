#!/usr/bin/env bats
#
# This tests the googlecompute builder. The teardown function will
# delete any images with the text "packerbats" within the name.

load test_helper
fixtures builder-googlecompute

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

teardown() {
    gcloud compute --format='table[no-heading](name)' --project=${GC_PROJECT_ID} images list \
        | grep packerbats \
        | xargs -n1 gcloud compute --project=${GC_PROJECT_ID} images delete
}

@test "googlecompute: build minimal.json" {
    run packer build ${USER_VARS} $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ "$(gc_has_image "packerbats-minimal")" -eq 1 ]
}
