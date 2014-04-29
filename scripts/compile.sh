#!/bin/bash
#
# This script compiles Packer for various platforms (specified by the
# PACKER_OS and PACKER_ARCH environmental variables).
set -e

NO_COLOR="\x1b[0m"
OK_COLOR="\x1b[32;01m"
ERROR_COLOR="\x1b[31;01m"
WARN_COLOR="\x1b[33;01m"

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd $DIR

# Get the git commit
GIT_COMMIT=$(git rev-parse HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "+CHANGES" || true)

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-linux darwin windows freebsd openbsd}

# Make sure that if we're killed, we kill all our subprocseses
trap "kill 0" SIGINT SIGTERM EXIT

echo -e "${OK_COLOR}==> Installing dependencies to speed up builds...${NO_COLOR}"
go get ./...

echo -e "${OK_COLOR}==> Beginning compile...${NO_COLOR}"
rm -rf pkg/
gox \
    -os="${XC_OS}" \
    -arch="${XC_ARCH}" \
    -ldflags "-X github.com/mitchellh/packer/packer.GitCommit ${GIT_COMMIT}${GIT_DIRTY}" \
    -output "pkg/{{.OS}}_{{.Arch}}/packer-{{.Dir}}" \
    ./...

# Make sure "packer-packer" is renamed properly
for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    set +e
    mv ${PLATFORM}/packer-packer*.exe ${PLATFORM}/packer.exe 2>/dev/null
    mv ${PLATFORM}/packer-packer* ${PLATFORM}/packer 2>/dev/null
    set -e
done

# Reset signal trapping to avoid "Terminated: 15" at the end
trap - SIGINT SIGTERM EXIT
