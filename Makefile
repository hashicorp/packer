all:
	@mkdir -p bin/
	go build -o bin/packer packer/bin-packer

.PHONY: all
