all:
	@mkdir -p bin/
	go get -d
	go build -a -o bin/packer

.PHONY: all
