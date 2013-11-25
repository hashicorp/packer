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

# Compile the thing
export XC_ARCH=$(go env GOARCH)
export XC_OS=$(go env GOOS)
./scripts/compile.sh

# Move all the compiled things to the PATH
cp pkg/${XC_OS}_${XC_ARCH}/* ${GOPATH}/bin
