#!/bin/sh

set -eux

export PACKER_ACC=1

go test -v -count 1 -timeout 20m ./driver ./iso ./clone
