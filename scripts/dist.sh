#!/usr/bin/env bash
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
    echo "Please specify version"
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

echo "==> Push with hc-releases"