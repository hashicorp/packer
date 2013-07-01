NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

all:
	@mkdir -p bin/
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -d -v ./...
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	@./scripts/build.sh

format:
	go fmt ./...

test:
	@echo "$(OK_COLOR)==> Testing Packer...$(NO_COLOR)"
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d
	go test ./...

.PHONY: all format test
