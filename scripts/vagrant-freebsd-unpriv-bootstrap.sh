#!/bin/sh
# Copyright IBM Corp. 2013, 2025
# SPDX-License-Identifier: BUSL-1.1


export GOPATH=/opt/gopath

PATH=$GOPATH/bin:$PATH
export PATH

cd /opt/gopath/src/github.com/hashicorp/packer && gmake deps
