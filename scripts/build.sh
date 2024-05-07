#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


# This script builds the application from source for multiple platforms.
# Determine the arch/os combos we're building for
ALL_XC_ARCH="386 amd64 arm arm64 ppc64le mips mips64 mipsle mipsle64 s390x"
ALL_XC_OS="linux darwin windows freebsd openbsd solaris"
SKIPPED_OSARCH="!darwin/arm !freebsd/arm !freebsd/arm64"

# Exit immediately if a command fails
set -e

# Validates that a necessary tool is on the PATH
function validateToolPresence
{
    local TOOLNAME=$1
    if ! which ${TOOLNAME} >/dev/null; then
        echo "${TOOLNAME} is not on the path. Exiting..."
        exit 1
    fi
}

# Validates that all used tools are present; exits when any is not found
function validatePreconditions
{
    echo "==> Checking for necessary tools..."
    validateToolPresence realpath
    validateToolPresence dirname
    validateToolPresence tr
    validateToolPresence find
}

# Get the parent directory of where this script is.
# NOTE: I'm unsure why you don't just use realpath like below
function enterPackerSourceDir
{
    echo "==> Entering Packer source dir..."
    local BUILD_SCRIPT_PATH="${BASH_SOURCE[0]}"
    SOURCEDIR=$(dirname $(dirname $(realpath "${BUILD_SCRIPT_PATH}")))
    cd ${SOURCEDIR}
}

function ensureOutputStructure {
    echo "==> Ensuring output directories are present..."
    mkdir -p bin/
    mkdir -p pkg/
}

function cleanOutputDirs {
    echo "==> Removing old builds..."
    rm -f bin/*
    rm -fr pkg/*
}

function lowerCaseOSType {
    local OS_TYPE=${OSTYPE:=`uname`}
    echo "${OS_TYPE}" | tr "[:upper:]" "[:lower:]"
}

# Returns the OS appropriate path separator
function getPathSeparator {
    # helpers for Cygwin-hosted builds
    case "$(lowerCaseOSType)" in
        mingw*|msys*|cygwin*)
            # cygwin only translates ';' to ':' on select environment variables
            echo ';'
            ;;
        *)	echo ':'
    esac
}

function convertPathOnCygwin() {
    local flag
    local somePath
    if [ "${1:0:1}" = '-' ]; then
        flag=$1
        somePath=$2
    else
        somePath=$1
    fi

    [ -n "${somePath}" ] || return 0
    case "$(lowerCaseOSType)" in
        cygwin*)
            cygpath ${flag} -- "${somePath}"
            ;;
        *)  echo "${somePath}"
    esac
}

validatePreconditions
enterPackerSourceDir
ensureOutputStructure
cleanOutputDirs

PATHSEP=$(getPathSeparator)

# XXX works in MINGW?
# FIXME: What if go is not in the PATH and GOROOT isn't set?
which go &>/dev/null || PATH+=":`convertPathOnCygwin "${GOROOT:?}"`/bin"

OLDIFS="${IFS}"

# make sure GOPATH is consistent - Windows binaries can't handle Cygwin-style paths
IFS="${PATHSEP}"
for d in ${GOPATH:-$(go env GOPATH)}; do
    _GOPATH+="${_GOPATH:+${PATHSEP}}$(convertPathOnCygwin --windows "${d}")"
done
GOPATH="$_GOPATH"

# locate 'gox' and traverse GOPATH if needed
which "${GOX:=gox}" &>/dev/null || {
    for d in ${GOPATH}; do
        GOX="$(convertPathOnCygwin --unix "${d}")/bin/gox"
        [ -x "${GOX}" ] && break || unset GOX
    done
}
IFS="$OLDIFS"

# Build!
echo "==> Building..."

# If in dev mode, only build for ourself
if [ -n "${PACKER_DEV+x}" ]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

export CGO_ENABLED=0

${GOX:?command not found} \
    -os="${XC_OS:-$ALL_XC_OS}" \
    -arch="${XC_ARCH:-$ALL_XC_ARCH}" \
    -osarch="${SKIPPED_OSARCH}" \
    -ldflags "${GOLDFLAGS}" \
    -output "pkg/{{.OS}}_{{.Arch}}/packer" \
    .

# trim GOPATH to first element
IFS="${PATHSEP}"
# FIXME: How do you know that the first path of GOPATH is the main GOPATH? Or is the main GOPATH meant to be the first path in GOPATH?
MAIN_GOPATH=(${GOPATH})
MAIN_GOPATH="$(convertPathOnCygwin --unix "${MAIN_GOPATH[0]}")"
IFS="${OLDIFS}"

# Copy our OS/Arch to the bin/ directory
echo "==> Copying binaries for this platform..."
DEV_PLATFORM="./pkg/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
    cp -v ${F} bin/
    cp -v ${F} "${MAIN_GOPATH}/bin/"
done

# Done!
echo
echo "==> Results:"
ls -hl bin/
