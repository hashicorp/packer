TEST?=$(shell go list ./... | grep -v vendor)
# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)
# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)

default: deps generate test dev

ci: deps test

release: deps test releasebin

bin: deps
	@echo "WARN: 'make bin' is for debug / test builds only. Use 'make release' for release builds."
	@sh -c "$(CURDIR)/scripts/build.sh"

releasebin: deps
	@grep 'const VersionPrerelease = ""' version.go > /dev/null ; if [ $$? -ne 0 ]; then \
		echo "ERROR: You must remove prerelease tags from version.go prior to release."; \
		exit 1; \
	fi
	@sh -c "$(CURDIR)/scripts/build.sh"

deps:
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/stringer
	go get golang.org/x/tools/cmd/vet
	@echo "INFO: Packer deps are managed by godep. See CONTRIBUTING.md"

dev: deps
	@grep 'const VersionPrerelease = ""' version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must add prerelease tags to version.go prior to making a dev build."; \
		exit 1; \
	fi
	@PACKER_DEV=1 sh -c "$(CURDIR)/scripts/build.sh"

# generate runs `go generate` to build the dynamically generated
# source files.
generate: deps
	go generate .
	go fmt command/plugin.go

test: deps
	@echo "INFO: Test results going to packer-test.log; this may take awhile"
	@go test $(TEST) $(TESTARGS) -timeout=15s | tee packer-test.log
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

# testacc runs acceptance tests
testacc: deps generate
	@echo "WARN: Acceptance tests will take a long time to run and may cost money. Ctrl-C if you want to cancel."
	PACKER_ACC=1 go test -v $(TEST) $(TESTARGS) -timeout=45m | tee packer-test-acc.log

testrace: deps
	@echo "INFO: Test results going to packer-test-race.log; this may take awhile"
	@go test -race $(TEST) $(TESTARGS) -timeout=15s | tee packer-test-race.log

updatedeps:
	go get -u github.com/mitchellh/gox
	go get -u golang.org/x/tools/cmd/stringer
	go get -u golang.org/x/tools/cmd/vet
	@echo "INFO: Packer deps are managed by godep. See CONTRIBUTING.md"

.PHONY: bin checkversion ci default deps generate releasebin test testacc testrace updatedeps
