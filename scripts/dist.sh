#!/bin/bash
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that dir because we expect that
cd $DIR

# Get the version from the command line
VERSION=$1
if [ -z $VERSION ]; then
    echo "Please specify a version."
    exit 1
fi

# Make sure we have a bintray API key
if [ -z $BINTRAY_API_KEY ]; then
    echo "Please set your bintray API key in the BINTRAY_API_KEY env var."
    exit 1
fi

# Zip and copy to the dist dir
echo "==> Packaging..."
rm -rf ./pkg/dist
mkdir -p ./pkg/dist
for PLATFORM in $(find ./pkg -mindepth 1 -maxdepth 1 -type d); do
    OSARCH=$(basename ${PLATFORM})

    if [ $OSARCH = "dist" ]; then
        continue
    fi

    echo "--> ${OSARCH}"
    pushd $PLATFORM >/dev/null 2>&1
    zip ../dist/packer_${VERSION}_${OSARCH}.zip ./*
    popd >/dev/null 2>&1
done

# Make the checksums
echo "==> Checksumming..."
pushd ./pkg/dist >/dev/null 2>&1
shasum -a256 * > ./packer_${VERSION}_SHA256SUMS
popd >/dev/null 2>&1

echo "==> Uploading..."
for ARCHIVE in ./pkg/dist/*; do
    ARCHIVE_NAME=$(basename ${ARCHIVE})

    echo Uploading: $ARCHIVE_NAME
    curl \
        -T ${ARCHIVE} \
        -umitchellh:${BINTRAY_API_KEY} \
        "https://api.bintray.com/content/mitchellh/packer/packer/${VERSION}/${ARCHIVE_NAME}"
done
