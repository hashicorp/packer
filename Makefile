TEST?=$(shell go list ./... | grep -v vendor)
VET?=$(shell ls -d */ | grep -v vendor | grep -v website)
# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)
# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)

default: deps generate test dev

ci: deps test

release: deps test releasebin package ## Build a release build

bin: deps ## Build debug/test build
	@echo "WARN: 'make bin' is for debug / test builds only. Use 'make release' for release builds."
	@GO15VENDOREXPERIMENT=1 sh -c "$(CURDIR)/scripts/build.sh"

releasebin: deps
	@grep 'const VersionPrerelease = "dev"' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must remove prerelease tags from version/version.go prior to release."; \
		exit 1; \
	fi
	@GO15VENDOREXPERIMENT=1 sh -c "$(CURDIR)/scripts/build.sh"

package:
	$(if $(VERSION),,@echo 'VERSION= needed to release; Use make package skip compilation'; exit 1)
	@sh -c "$(CURDIR)/scripts/dist.sh $(VERSION)"

deps:
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/stringer
	@go version | grep 1.4 ; if [ $$? -eq 0 ]; then \
		echo "Installing godep and restoring dependencies"; \
		go get github.com/tools/godep; \
		godep restore; \
	fi

dev: deps ## Build and install a development build
	@grep 'const VersionPrerelease = ""' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must add prerelease tags to version/version.go prior to making a dev build."; \
		exit 1; \
	fi
	@PACKER_DEV=1 GO15VENDOREXPERIMENT=1 sh -c "$(CURDIR)/scripts/build.sh"

fmt: ## Format Go code
	go fmt `go list ./... | grep -v vendor`

# Install js-beautify with npm install -g js-beautify
fmt-examples:
	find examples -name *.json | xargs js-beautify -r -s 2 -n -eol "\n"

# generate runs `go generate` to build the dynamically generated
# source files.
generate: deps ## Generate dynamically generated code
	go generate .
	go fmt command/plugin.go

test: deps ## Run unit tests
	@go test $(TEST) $(TESTARGS) -timeout=2m
	@go tool vet $(VET)  ; if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

# testacc runs acceptance tests
testacc: deps generate ## Run acceptance tests
	@echo "WARN: Acceptance tests will take a long time to run and may cost money. Ctrl-C if you want to cancel."
	PACKER_ACC=1 go test -v $(TEST) $(TESTARGS) -timeout=45m

testrace: deps ## Test for race conditions
	@go test -race $(TEST) $(TESTARGS) -timeout=2m

updatedeps:
	go get -u github.com/mitchellh/gox
	go get -u golang.org/x/tools/cmd/stringer
	@echo "INFO: Packer deps are managed by godep. See CONTRIBUTING.md"

# This is used to add new dependencies to packer. If you are submitting a PR
# that includes new dependencies you will need to run this.
vendor: ## Add new dependencies.
	godep restore
	godep save

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: bin checkversion ci default deps fmt fmt-examples generate releasebin test testacc testrace updatedeps
