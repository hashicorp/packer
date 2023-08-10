#! /usr/bin/env bash
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1


set -eu -o pipefail

gpath=${GOPATH:-}
if [ -z "$gpath" ]; then
  gpath=$HOME/go
fi

reldir=`dirname $0`
curdir=`pwd`
cd $reldir
CUR_GO_DIR=`pwd`
cd $curdir

if [[ $CUR_GO_DIR == *"$gpath"* ]]; then
  # echo "You're on the gopath"
  exit 1
else
  # echo "You're not on the gopath"
  exit 0
fi