#!/usr/bin/env bash

# Get the full path to the directory where this script is, because
# GOPATH prefers full paths.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )" && pwd )"

# Setup our GOPATH
echo "Setting GOPATH to: ${DIR}"
export GOPATH="${DIR}"
