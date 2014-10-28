TEST?=./...

default: test

bin:
	@sh -c "$(CURDIR)/scripts/build.sh"

dev:
	@TF_DEV=1 sh -c "$(CURDIR)/scripts/build.sh"

test:
	go test $(TEST) $(TESTARGS) -timeout=10s

testrace:
	go test -race $(TEST) $(TESTARGS)

updatedeps:
	go get -d -v -p 2 ./...
	@sh -c "go get -u -v -p 2 ./...; exit 0"
	go get -d -v -p 2 ./...

.PHONY: bin default test updatedeps
