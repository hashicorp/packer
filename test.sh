#!/bin/sh

set -eux

go test -c ./driver
go test -c ./iso
go test -c ./clone

export PACKER_ACC=1
go test -v ./driver ./iso ./clone
