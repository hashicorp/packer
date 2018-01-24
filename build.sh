#!/bin/sh

set -eux

glide install -v
export CGO_ENABLED=0
export GOARCH=amd64
mkdir -p bin
rm -f bin/*

GOOS=darwin  go build -o bin/packer-builder-vsphere-clone.macos ./clone
GOOS=linux   go build -o bin/packer-builder-vsphere-clone.linux ./clone
GOOS=windows go build -o bin/packer-builder-vsphere-clone.exe   ./clone

GOOS=darwin  go build -o bin/packer-builder-vsphere-iso.macos ./iso
GOOS=linux   go build -o bin/packer-builder-vsphere-iso.linux ./iso
GOOS=windows go build -o bin/packer-builder-vsphere-iso.exe   ./iso
