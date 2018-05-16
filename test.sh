#!/bin/sh

set -eux

export PACKER_ACC=1

go test -v ./driver ./iso ./clone
