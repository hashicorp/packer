#!/bin/sh
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

# Compile the main Packer app
echo "${OK_COLOR}--> Compiling Packer${NO_COLOR}"
go build -v -o bin/packer .

# Go over each plugin and build it
for PLUGIN in $(find ./plugin -mindepth 1 -maxdepth 1 -type d); do
    PLUGIN_NAME=$(basename ${PLUGIN})
    echo "${OK_COLOR}--> Compiling Plugin: ${PLUGIN_NAME}${NO_COLOR}"
    go build -v -o bin/packer-${PLUGIN_NAME} ${PLUGIN}
done
