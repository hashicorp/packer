#!/bin/sh

set -eux

glide install -v
export CGO_ENABLED=0
export GOARCH=amd64
mkdir -p bin
rm -f bin/*

GOOS=darwin  go build -o bin/packer-builder-vsphere.macos ./cmd/clone
GOOS=linux   go build -o bin/packer-builder-vsphere.linux ./cmd/clone
GOOS=windows go build -o bin/packer-builder-vsphere.exe   ./cmd/clone

GOOS=darwin  go build -o bin/packer-builder-vsphere-iso.macos ./cmd/iso
GOOS=linux   go build -o bin/packer-builder-vsphere-iso.linux ./cmd/iso
GOOS=windows go build -o bin/packer-builder-vsphere-iso.exe   ./cmd/iso
