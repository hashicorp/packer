#!/bin/bash
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that
cd $DIR

# Determine the version that we're building based on the contents
# of packer/version.go.
VERSION=$(grep "const Version " packer/version.go | sed -E 's/.*"(.+)"$/\1/')
VERSIONDIR="${VERSION}"
PREVERSION=$(grep "const VersionPrerelease " packer/version.go | sed -E 's/.*"(.*)"$/\1/')
if [ ! -z $PREVERSION ]; then
    PREVERSION="${PREVERSION}.$(date -u +%s)"
    VERSIONDIR="${VERSIONDIR}-${PREVERSION}"
fi

# This function waits for all background tasks to complete
waitAll() {
    RESULT=0
    for job in `jobs -p`; do
        wait $job
        if [ $? -ne 0 ]; then
            RESULT=1
        fi
    done

    if [ $RESULT -ne 0 ]; then
        exit $RESULT
    fi
}

# Compile the main project
./scripts/compile.sh

# Make sure that if we're killed, we kill all our subprocseses
trap "kill 0" SIGINT SIGTERM EXIT

# Zip all the packages
mkdir -p ./pkg/dist
for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    PLATFORM_NAME=$(basename ${PLATFORM})
    ARCHIVE_NAME="${VERSIONDIR}_${PLATFORM_NAME}"

    if [ $PLATFORM_NAME = "dist" ]; then
        continue
    fi

    (
    pushd ${PLATFORM}
    zip ${DIR}/pkg/dist/${ARCHIVE_NAME}.zip ./*
    popd
    ) &
done

waitAll

# Make the checksums
pushd ./pkg/dist
shasum -a256 * > ./${VERSIONDIR}_SHA256SUMS
popd

exit 0
