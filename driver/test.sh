#!/bin/sh

export VSPHERE_DRIVER_ACC=1
go test -v "$@"
