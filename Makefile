all:
	@mkdir -p bin/
	go get -d ./...
	go build -a -o bin/packer

test:
	go test ./...

.PHONY: all test
