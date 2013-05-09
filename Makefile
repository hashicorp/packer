NO_COLOR=\x1b[0m
OK_COLOR=\x1b[32;01m
ERROR_COLOR=\x1b[31;01m
WARN_COLOR=\x1b[33;01m

export ROOTDIR=$(CURDIR)

all:
	@mkdir -p bin/
	go get -d -v ./...
	@echo "$(OK_COLOR)--> Compiling Packer...$(NO_COLOR)"
	go build -v -o bin/packer
	@echo "$(OK_COLOR)--> Compiling Builder: Amazon EBS...$(NO_COLOR)"
	$(MAKE) -C plugin/builder-amazon-ebs
	@echo "$(OK_COLOR)--> Compiling Command: Build...$(NO_COLOR)"
	$(MAKE) -C plugin/command-build

format:
	go fmt ./...

test:
	@echo "$(OK_COLOR)--> Testing Packer...$(NO_COLOR)"
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | xargs -n1 go get -d
	go test ./...

.PHONY: all format test
