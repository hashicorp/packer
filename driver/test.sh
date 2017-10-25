#!/bin/sh

export VSPHERE_DRIVER_ACC=1
cd testing
go test -v "$@"
