#!/usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

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

# Tag, unless told not to
if [ -z $NOTAG ]; then
  echo "==> Tagging..."
  git commit --allow-empty -a --gpg-sign=348FFC4C -m "Cut version $VERSION"
  git tag -a -m "Version $VERSION" -s -u 348FFC4C "v${VERSION}" $RELBRANCH
fi

# Zip all the files
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

./scripts/sign.sh
if [ $? -ne 0 ]; then
  exit 1
fi

if [ -z $NOSIGN ]; then
  echo "==> Signing..."
  pushd ./pkg/dist
  rm -f ./packer_${VERSION}_SHA256SUMS*
  shasum -a256 * > ./packer_${VERSION}_SHA256SUMS
  gpg --default-key 348FFC4C --detach-sig ./packer_${VERSION}_SHA256SUMS
  popd
fi

hc-releases upload $DIR/pkg/dist/
hc-releases publish

exit 0
