#!/bin/bash
#
# This script only builds the application from source.
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

# If we're building on Windows, specify an extension
EXTENSION=""
if [ "$(go env GOOS)" = "windows" ]; then
    EXTENSION=".exe"
fi

# Make sure that if we're killed, we kill all our subprocseses
trap "kill 0" SIGINT SIGTERM EXIT

# If we're building a race-enabled build, then set that up.
if [ ! -z $PACKER_RACE ]; then
    echo -e "${OK_COLOR}--> Building with race detection enabled${NO_COLOR}"
    PACKER_RACE="-race"
fi

echo -e "${OK_COLOR}--> Installing dependencies to speed up builds...${NO_COLOR}"
go get ./...

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

waitSingle() {
    if [ ! -z $PACKER_NO_BUILD_PARALLEL ]; then
        waitAll
    fi
}

if [ -z $PACKER_NO_BUILD_PARALLEL ]; then
    echo -e "${OK_COLOR}--> NOTE: Compilation of components " \
        "will be done in parallel.${NO_COLOR}"
fi

# Compile the main Packer app
echo -e "${OK_COLOR}--> Compiling Packer${NO_COLOR}"
(
go build \
    ${PACKER_RACE} \
    -ldflags "-X github.com/mitchellh/packer/packer.GitCommit ${GIT_COMMIT}${GIT_DIRTY}" \
    -v \
    -o bin/packer${EXTENSION} .

    cp bin/packer${EXTENSION} ${GOPATH}/bin
) &

waitSingle

# Go over each plugin and build it
for PLUGIN in $(find ./plugin -mindepth 1 -maxdepth 1 -type d); do
    PLUGIN_NAME=$(basename ${PLUGIN})
    echo -e "${OK_COLOR}--> Compiling Plugin: ${PLUGIN_NAME}${NO_COLOR}"
    (
    go build \
        ${PACKER_RACE} \
        -ldflags "-X github.com/mitchellh/packer/packer.GitCommit ${GIT_COMMIT}${GIT_DIRTY}" \
        -v \
        -o bin/packer-${PLUGIN_NAME}${EXTENSION} ${PLUGIN}

        cp bin/packer-${PLUGIN_NAME}${EXTENSION} ${GOPATH}/bin
    ) &

    waitSingle
done

waitAll

# Reset signal trapping to avoid "Terminated: 15" at the end
trap - SIGINT SIGTERM EXIT
