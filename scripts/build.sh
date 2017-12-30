#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd $DIR

: ${GOPATH:=$(go env GOPATH)}
[ -n "$GOPATH" ] || { echo "Error: GOPATH not set"; exit 1; }

# Windows version of 'go' tools can not cope with Cygwin style paths
case $(uname) in
    CYGWIN*)
	GOX="$(cygpath -u "$GOPATH")/bin/gox"
	GOPATH=$(cygpath -w "$GOPATH")
        ;;
esac

# If its dev mode, only build for ourself
if [ -n "${PACKER_DEV}" ]; then
    : ${XC_OS:=$(go env GOOS)}
    : ${XC_ARCH:=$(go env GOARCH)}
fi

# Determine the arch/os combos we're building for
: ${XC_ARCH:="386 amd64 arm arm64 ppc64le"}
: ${XC_OS:="linux darwin windows freebsd openbsd solaris"}

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

OLDIFS=$IFS
IFS=:
case $(uname) in
    MINGW*|MSYS*)
        IFS=";"
        ;;
esac

# Build!
echo "==> Building..."
set +e
${GOX:-$GOPATH/bin/gox} \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -osarch="!darwin/arm !darwin/arm64" \
    -ldflags "${GOLDFLAGS}" \
    -output "pkg/{{.OS}}_{{.Arch}}/packer" \
    .

IFS=$OLDIFS
set -e

# Copy our OS/Arch to the bin/ directory
echo "==> Copying binaries for this platform..."
for F in $(find pkg/$(go env GOOS)_$(go env GOARCH) -mindepth 1 -maxdepth 1 -type f); do
    cp -v ${F} bin/
    cp -v ${F} ${GOPATH}/bin/
done

# Done!
echo
echo "==> Results:"
ls -hl bin/
