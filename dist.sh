#!/bin/bash
set -e

# Get the directory where this script is. This will also resolve
# any symlinks in the directory/script, so it will be the fully
# resolved path.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# Determine the version that we're building based on the contents
# of packer/version.go.
VERSION=$(grep "const Version " packer/version.go | sed -E 's/.*"(.+)"$/\1/')
PREVERSION=$(grep "const VersionPrerelease " packer/version.go | sed -E 's/.*"(.+)"$/\1/')
if [ ! -z $PREVERSION ]; then
    PREVERSION="${PREVERSION}.$(date -u +%s)"
fi

echo "Version: ${VERSION} ${PREVERSION}"

# This function builds whatever directory we're in...
xc() {
    goxc \
        -arch="386 amd64 arm" \
        -os="linux darwin windows freebsd openbsd" \
        -d="${DIR}/pkg" \
        -pv="${VERSION}" \
        -pr="${PREVERSION}" \
        go-install \
        xc
}

# Build our root project
xc

# Build all the plugins
for PLUGIN in $(find ./plugin -mindepth 1 -maxdepth 1 -type d); do
    PLUGIN_NAME=$(basename ${PLUGIN})
    pushd ${PLUGIN}
    xc
    popd
    find ./pkg -type f -name ${PLUGIN_NAME} -execdir mv ${PLUGIN_NAME} packer-${PLUGIN_NAME} ';'
done
