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
	(go get -u -v -p 2 ./...; exit 0)
	cd ../../rackspace/gophercloud; git checkout release/v0.1.1  #TODO: goddammit dependency management

.PHONY: bin default test updatedeps
