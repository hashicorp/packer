all:
	@mkdir -p bin/
	go get
	go build -o bin/packer

.PHONY: all
