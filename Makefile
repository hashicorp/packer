NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

all: deps
	@mkdir -p bin/
	@echo "$(OK_COLOR)==> Building$(NO_COLOR)"
	@./scripts/build.sh

deps:
	@echo "$(OK_COLOR)==> Installing dependencies$(NO_COLOR)"
	@go get -d -v -u ./...
	@go list -f '{{range .TestImports}}{{.}} {{end}}' ./... | xargs -n1 go get -d

clean:
	@rm -rf bin/ local/ pkg/ src/ website/.sass-cache website/build

format:
	go fmt ./...

test: deps
	@echo "$(OK_COLOR)==> Testing Packer...$(NO_COLOR)"
	go test ./...

.PHONY: all deps format test
