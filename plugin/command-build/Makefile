plugin:
	go get -d -v ./...
	go build -v -o $(ROOTDIR)/bin/packer-build

format:
	go fmt ./...

test:
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | xargs -n1 go get -d
	go test ./...

.PHONY: all format test
