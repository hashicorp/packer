all:
	@mkdir -p bin/
	@rm -r ${GOPATH}/pkg
	go get -d -v ./...
	go build -v -o bin/packer

format:
	go fmt ./...

test:
	@go list -f '{{range .TestImports}}{{.}}\
		{{end}}' ./... | xargs -n1 go get -d
	@go test -i ./...
	go test ./...

.PHONY: all format test
