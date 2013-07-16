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

echo "Version: ${VERSION} ${PREVERSION}"

# Determine the arch/os combos we're building for
XC_ARCH=${XC_ARCH:-"386 amd64 arm"}
XC_OS=${XC_OS:-linux darwin windows freebsd openbsd}

echo "Arch: ${XC_ARCH}"
echo "OS: ${XC_OS}"

# This function builds whatever directory we're in...
xc() {
    goxc \
        -arch="$XC_ARCH" \
        -os="$XC_OS" \
        -d="${DIR}/pkg" \
        -pv="${VERSION}" \
        -pr="${PREVERSION}" \
        go-install \
        xc
}

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

# Make sure that if we're killed, we kill all our subprocseses
trap "kill 0" SIGINT SIGTERM EXIT

# Build our root project
xc &

# Build all the plugins
for PLUGIN in $(find ./plugin -mindepth 1 -maxdepth 1 -type d); do
    PLUGIN_NAME=$(basename ${PLUGIN})
    (
    pushd ${PLUGIN}
    xc
    popd
    find ./pkg \
        -type f \
        -name ${PLUGIN_NAME} \
        -execdir mv ${PLUGIN_NAME} packer-${PLUGIN_NAME} ';'
    find ./pkg \
        -type f \
        -name ${PLUGIN_NAME}.exe \
        -execdir mv ${PLUGIN_NAME}.exe packer-${PLUGIN_NAME}.exe ';'
    ) &
done

waitAll

# Zip all the packages
mkdir -p ./pkg/${VERSIONDIR}/dist
for PLATFORM in $(find ./pkg/${VERSIONDIR} -mindepth 1 -maxdepth 1 -type d); do
    PLATFORM_NAME=$(basename ${PLATFORM})
    ARCHIVE_NAME="${VERSIONDIR}_${PLATFORM_NAME}"

    if [ $PLATFORM_NAME = "dist" ]; then
        continue
    fi

    (
    pushd ${PLATFORM}
    zip ${DIR}/pkg/${VERSIONDIR}/dist/${ARCHIVE_NAME}.zip ./*
    popd
    ) &
done

waitAll

# Make the checksums
pushd ./pkg/${VERSIONDIR}/dist
shasum -a256 * > ./${VERSIONDIR}_SHA256SUMS
popd

exit 0
