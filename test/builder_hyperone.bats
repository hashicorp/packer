#!/usr/bin/env bats
#
# This tests the hyperone builder. The teardown function will
# delete any images with the text "packerbats" within the name.

load test_helper
fixtures builder-hyperone

# Required parameters
: ${HYPERONE_TOKEN:?}
: ${HYPERONE_PROJECT:?}
command -v h1 >/dev/null 2>&1 || {
    echo "'h1' must be installed" >&2
    exit 1
}

USER_VARS="${USER_VARS} -var token=${HYPERONE_TOKEN}"
USER_VARS="${USER_VARS} -var project=${HYPERONE_PROJECT}"

hyperone_has_image() {
    h1 image list --project-select=${HYPERONE_PROJECT} --output=tsv | grep $1 | wc -l
}

teardown() {
    h1 image list --project-select=${HYPERONE_PROJECT} --output=tsv \
        | grep packerbats \
        | awk '{print $1}' \
        | xargs -n1 h1 image delete --project-select=${HYPERONE_PROJECT} --yes --image
}

@test "hyperone: build minimal.json" {
    run packer build ${USER_VARS} $FIXTURE_ROOT/minimal.json
    [ "$status" -eq 0 ]
    [ "$(hyperone_has_image "packerbats-minimal")" -eq 1 ]
}
