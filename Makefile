all:
	@mkdir -p bin/
	go build -o bin/packer

.PHONY: all
