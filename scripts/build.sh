#!/usr/bin/env bash

# This script builds the application from source for multiple platforms.
# Determine the arch/os combos we're building for
ALL_XC_ARCH="386 amd64 arm arm64 ppc64le"
ALL_XC_OS="linux darwin windows freebsd openbsd solaris"

set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd $DIR

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# helpers for Cygwin-hosted builds
: ${OSTYPE:=`uname`}

case $OSTYPE in
    MINGW*|MSYS*|cygwin|CYGWIN*)
        # cygwin only translates ';' to ':' on select environment variables
        PATHSEP=';'
        ;;
    *)	PATHSEP=':'
esac

function convert_path() {
    local flag
    [ "${1:0:1}" = '-' ] && { flag="$1"; shift; }

    [ -n "$1" ] || return 0
    case ${OSTYPE:-`uname`} in
        cygwin|CYGWIN*)
            cygpath $flag -- "$1"
            ;;
        *)  echo "$1"
    esac
}

# XXX works in MINGW?
which go &>/dev/null || PATH+=":`convert_path "${GOROOT:?}"`/bin"

OLDIFS="$IFS"

# make sure GOPATH is consistent - Windows binaries can't handle Cygwin-style paths
IFS="$PATHSEP"
for d in ${GOPATH:-$(go env GOPATH)}; do
    _GOPATH+="${_GOPATH:+$PATHSEP}$(convert_path --windows "$d")"
done
GOPATH="$_GOPATH"

# locate 'gox' and traverse GOPATH if needed
which "${GOX:=gox}" &>/dev/null || {
    for d in $GOPATH; do
        GOX="$(convert_path --unix "$d")/bin/gox"
        [ -x "$GOX" ] && break || unset GOX
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

set +e
${GOX:?command not found} \
    -os="${XC_OS:-$ALL_XC_OS}" \
    -arch="${XC_ARCH:-$ALL_XC_ARCH}" \
    -osarch="!darwin/arm !darwin/arm64" \
    -ldflags "${GOLDFLAGS}" \
    -output "pkg/{{.OS}}_{{.Arch}}/packer" \
    .
set -e

# trim GOPATH to first element
IFS="$PATHSEP"
MAIN_GOPATH=($GOPATH)
MAIN_GOPATH="$(convert_path --unix "$MAIN_GOPATH")"
IFS=$OLDIFS

# Copy our OS/Arch to the bin/ directory
echo "==> Copying binaries for this platform..."
DEV_PLATFORM="./pkg/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f 2>/dev/null); do
    cp -v ${F} bin/
    cp -v ${F} ${MAIN_GOPATH}/bin/
done

# Done!
echo
echo "==> Results:"
ls -hl bin/
