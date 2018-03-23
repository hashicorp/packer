#!/bin/sh

set -eux

dep ensure
export CGO_ENABLED=0
export GOARCH=amd64
mkdir -p bin
rm -f bin/*

GOOS=darwin  go build -o bin/packer-builder-vsphere-iso.macos ./cmd/iso
GOOS=linux   go build -o bin/packer-builder-vsphere-iso.linux ./cmd/iso
GOOS=windows go build -o bin/packer-builder-vsphere-iso.exe   ./cmd/iso

GOOS=darwin  go build -o bin/packer-builder-vsphere-clone.macos ./cmd/clone
GOOS=linux   go build -o bin/packer-builder-vsphere-clone.linux ./cmd/clone
GOOS=windows go build -o bin/packer-builder-vsphere-clone.exe   ./cmd/clone
