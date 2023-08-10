# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


#!/usr/bin/env bash

# This script uploads the Darwin builds to artifactory, then triggers the
# circle ci job that signs them.

# ARTIFACTORY_USER="sa-circle-codesign"
# export PRODUCT_NAME="packer"
# export ARTIFACTORY_TOKEN=$ARTIFACTORY_TOKEN

ARTIFACTORY_TOKEN="${ARTIFACTORY_TOKEN:-""}"

if [ -z "$ARTIFACTORY_TOKEN" ]; then
  echo "Missing required Artifactory credentials"
  exit 1
fi

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"
# Change into that dir because we expect that
cd $DIR

BIN_UUIDS=()
BUILD_NUMBERS=()
for DARWIN_BIN in $(find ./pkg/dist/*darwin_*.zip); do
  echo "signing $DARWIN_BIN"
  export ARTIFACTORY_USER="sa-circle-codesign"
  export PRODUCT_NAME="packer"
  export ARTIFACTORY_TOKEN=$ARTIFACTORY_TOKEN
  export TARGET_ZIP=$DARWIN_BIN

  echo $TARGET_ZIP
  ./scripts/codesign_example.sh
  if [ $? -ne 0 ]; then
  	exit 1
  fi
done

exit 0
