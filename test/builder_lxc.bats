#!/usr/bin/env bats
#
# This tests the lxc builder by creating minimal containers and checking that
# custom lxc container configuration files are successfully applied. The
# teardown function will delete any images in the output-lxc-* folders along
# with the auto-generated lxc container configuration files and hook scripts.

#load test_helper
#fixtures builder-lxc
FIXTURE_ROOT="$BATS_TEST_DIRNAME/fixtures/builder-lxc"

have_command() {
    command -v "$1" >/dev/null 2>&1
}

# Required parameters
have_command lxc-create || {
    echo "'lxc-create' must be installed via the lxc (or lxc1 for ubuntu >=16.04) package" >&2
    exit 1
}

DESTROY_HOOK_SCRIPT=$FIXTURE_ROOT/destroy-hook.sh
DESTROY_HOOK_LOG=$FIXTURE_ROOT/destroy-hook.log
printf > "$DESTROY_HOOK_SCRIPT" '
echo "$LXC_NAME" > "%s"
' "$DESTROY_HOOK_LOG"
chmod +x "$DESTROY_HOOK_SCRIPT"

INIT_CONFIG=$FIXTURE_ROOT/lxc.custom.conf
printf > "$INIT_CONFIG" '
lxc.hook.destroy = %s
' "$DESTROY_HOOK_SCRIPT"

teardown() {
    for f in "$INIT_CONFIG" "$DESTROY_HOOK_SCRIPT" "$DESTROY_HOOK_LOG"; do
        [ -e "$f" ] && rm -f "$f"
    done

    rm -rf output-lxc-*
}

assert_build() {
    local template_name="$1"
    shift

    local build_status=0

    run packer build -var template_name="$template_name"  "$@"

    [ "$status" -eq 0 ] || {
        echo "${template_name} build exited badly: $status" >&2
        echo "$output" >&2
        build_status="$status"
    }

    for expected in "output-lxc-${template_name}"/{rootfs.tar.gz,lxc-config}; do
        [ -f "$expected" ] || {
            echo "missing expected artifact '${expected}'" >&2
            build_status=1
        }
    done

    return $build_status
}

assert_container_name() {
    local container_name="$1"

    [ -f "$DESTROY_HOOK_LOG" ] || {
        echo "missing expected lxc.hook.destroy logfile '$DESTROY_HOOK_LOG'"
        return 1
    }

    read -r lxc_name < "$DESTROY_HOOK_LOG"

    [ "$lxc_name" = "$container_name" ]
}

@test "lxc: build centos minimal.json" {
    have_command yum || skip "'yum' must be installed to build centos containers"
    local container_name=packer-lxc-centos
    assert_build centos -var init_config="$INIT_CONFIG" \
        -var container_name="$container_name" \
        $FIXTURE_ROOT/minimal.json
    assert_container_name "$container_name"
}

@test "lxc: build trusty minimal.json" {
    have_command debootstrap || skip "'debootstrap' must be installed to build ubuntu containers"
    local container_name=packer-lxc-ubuntu
    assert_build ubuntu -var init_config="$INIT_CONFIG" \
        -var container_name="$container_name" \
        -var template_parameters="SUITE=trusty" \
        $FIXTURE_ROOT/minimal.json
    assert_container_name "$container_name"
}

@test "lxc: build debian minimal.json" {
    have_command debootstrap || skip "'debootstrap' must be installed to build debian containers"
    local container_name=packer-lxc-debian
    assert_build debian -var init_config="$INIT_CONFIG" \
        -var container_name="$container_name" \
        -var template_parameters="SUITE=jessie" \
        $FIXTURE_ROOT/minimal.json
    assert_container_name "$container_name"
}
