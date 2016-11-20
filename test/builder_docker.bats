#!/usr/bin/env bats
#
# This tests the docker builder.

load test_helper
fixtures builder-docker

# Required parameters
: ${DOCKER_USER:?}
: ${DOCKER_PASSWORD:?}
: ${DOCKER_REPOSITORY:?}
command -v docker info >/dev/null 2>&1 || {
    echo "'docker' must be working" >&2
    exit 1
}

USER_VARS="${USER_VARS} -var username=${DOCKER_USER}"
USER_VARS="${USER_VARS} -var password=${DOCKER_PASSWORD}"
USER_VARS="${USER_VARS} -var repository=${DOCKER_REPOSITORY}"

teardown() {
  rm -f alpine.tar
}

last_id() {
  docker images --format '{{.ID}}' | head -1
}

check_image() {
  [ "`docker inspect --format "$1" $(last_id)`" == "$2" ]
}

check_file_exists() {
  docker run $(last_id) ls $1
}

# Check if a file with full path $2 exists in tar file $1.
# No leading / in path.
check_file_exists_in_tar() {
  tar -tf $1 | egrep -q "^$2$"
}

@test "docker: build commit.json" {
    run packer build $USER_VARS $FIXTURE_ROOT/commit.json
    [ "$status" -eq 0 ]
    check_file_exists "/tmp/file"
}

@test "docker: build docker-tag.json" {
    run packer build $USER_VARS $FIXTURE_ROOT/docker-tag.json
    [ "$status" -eq 0 ]
    check_image '{{(index .RepoTags 0)}}' "${DOCKER_REPOSITORY}:latest"
}

@test "docker: build export.json" {
    run packer build $USER_VARS $FIXTURE_ROOT/export.json
    [ "$status" -eq 0 ]
    ls alpine.tar
}

@test "docker: build login.json" {
    run packer build $USER_VARS $FIXTURE_ROOT/login.json
    [ "$status" -eq 0 ]
    ls alpine.tar
}

@test "docker: build provisioner.json" {
    run packer build $USER_VARS $FIXTURE_ROOT/provisioner.json
    [ "$status" -eq 0 ]
    check_file_exists_in_tar alpine.tar tmp/metadata.txt
}
