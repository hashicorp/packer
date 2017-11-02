#!/bin/sh

export GOPATH=/opt/gopath

PATH=$GOPATH/bin:$PATH
export PATH

cd /opt/gopath/src/github.com/hashicorp/packer && gmake deps
