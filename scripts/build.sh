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

# Get the git commit
GIT_COMMIT=$(git rev-parse HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)

# If its dev mode, only build for ourself
if [ "${PACKER_DEV}x" != "x" ]; then
    XC_OS=${XC_OS:-$(go env GOOS)}
    XC_ARCH=${XC_ARCH:-$(go env GOARCH)}
fi

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-linux darwin windows freebsd openbsd}

# Delete the old dir
echo "==> Removing old directory..."
rm -f bin/*
rm -rf pkg/*
mkdir -p bin/

# Build!
echo "==> Building..."
set +e
gox \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -ldflags "-X github.com/mitchellh/packer/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY}" \
    -output "pkg/{{.OS}}_{{.Arch}}/packer" \
    .
set -e

# Move all the compiled things to the $GOPATH/bin
GOPATH=${GOPATH:-$(go env GOPATH)}
case $(uname) in
    CYGWIN*)
        GOPATH="$(cygpath $GOPATH)"
        ;;
esac
OLDIFS=$IFS
IFS=:
case $(uname) in
    MINGW*)
        IFS=";"
        ;;
    MSYS*)
        IFS=";"
        ;;
esac
MAIN_GOPATH=($GOPATH)
IFS=$OLDIFS

# Copy our OS/Arch to the bin/ directory
echo "==> Copying binaries for this platform..."
DEV_PLATFORM="./pkg/$(go env GOOS)_$(go env GOARCH)"
for F in $(find ${DEV_PLATFORM} -mindepth 1 -maxdepth 1 -type f); do
    cp ${F} bin/
    cp ${F} ${MAIN_GOPATH}/bin/
done

# Done!
echo
echo "==> Results:"
ls -hl bin/
