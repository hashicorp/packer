all:
	@mkdir -p bin/
	go get -a
	go build -a -o bin/packer

.PHONY: all
