GOOPTS := GOARCH=amd64 CGO_ENABLED=0

build: iso clone

iso: iso-linux iso-windows iso-macos
clone: clone-linux clone-windows clone-macos

iso-linux: modules bin
	$(GOOPTS) GOOS=linux go build -o bin/packer-builder-vsphere-iso.linux ./cmd/iso

iso-windows: modules bin
	$(GOOPTS) GOOS=windows go build -o bin/packer-builder-vsphere-iso.exe ./cmd/iso

iso-macos: modules bin
	$(GOOPTS) GOOS=darwin go build -o bin/packer-builder-vsphere-iso.macos ./cmd/iso

clone-linux: modules bin
	$(GOOPTS) GOOS=linux go build -o bin/packer-builder-vsphere-clone.linux ./cmd/clone

clone-windows: modules bin
	$(GOOPTS) GOOS=windows go build -o bin/packer-builder-vsphere-clone.exe ./cmd/clone

clone-macos: modules bin
	$(GOOPTS) GOOS=darwin go build -o bin/packer-builder-vsphere-clone.macos ./cmd/clone

modules:
	go mod download

bin:
	mkdir -p bin
	rm -f bin/*

test:
	PACKER_ACC=1 go test -v -count 1 ./driver ./iso ./clone

.PHONY: bin test
