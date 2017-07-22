TEST?=$(shell go list ./... | grep -v vendor)
VET?=$(shell ls -d */ | grep -v vendor | grep -v website)
# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)
# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)
GOFMT_FILES?=$$(find . -not -path "./vendor/*" -name "*.go")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)

# Get the git commit
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_IMPORT=github.com/hashicorp/packer/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)

export GOLDFLAGS

default: deps generate test dev

ci: deps test

release: deps test releasebin package ## Build a release build

bin: deps ## Build debug/test build
	@go get github.com/mitchellh/gox
	@echo "WARN: 'make bin' is for debug / test builds only. Use 'make release' for release builds."
	@sh -c "$(CURDIR)/scripts/build.sh"

releasebin: deps
	@go get github.com/mitchellh/gox
	@grep 'const VersionPrerelease = "dev"' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must remove prerelease tags from version/version.go prior to release."; \
		exit 1; \
	fi
	@sh -c "$(CURDIR)/scripts/build.sh"

package:
	$(if $(VERSION),,@echo 'VERSION= needed to release; Use make package skip compilation'; exit 1)
	@sh -c "$(CURDIR)/scripts/dist.sh $(VERSION)"

deps:
	@go get golang.org/x/tools/cmd/stringer
	@go get github.com/kardianos/govendor
	@govendor sync

dev: deps ## Build and install a development build
	@grep 'const VersionPrerelease = ""' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must add prerelease tags to version/version.go prior to making a dev build."; \
		exit 1; \
	fi
	@mkdir -p pkg/$(GOOS)_$(GOARCH)
	@go install -ldflags '$(GOLDFLAGS)'
	@cp $(GOPATH)/bin/packer bin
	@cp $(GOPATH)/bin/packer pkg/$(GOOS)_$(GOARCH)

fmt: ## Format Go code
	@gofmt -w -s $(GOFMT_FILES)

fmt-check: ## Check go code formatting
	$(CURDIR)/scripts/gofmtcheck.sh $(GOFMT_FILES)

fmt-docs:
	@find ./website/source/docs -name "*.md" -exec pandoc --wrap auto --columns 79 --atx-headers -s -f "markdown_github+yaml_metadata_block" -t "markdown_github+yaml_metadata_block" {} -o {} \;

# Install js-beautify with npm install -g js-beautify
fmt-examples:
	find examples -name *.json | xargs js-beautify -r -s 2 -n -eol "\n"

# generate runs `go generate` to build the dynamically generated
# source files.
generate: deps ## Generate dynamically generated code
	go generate .
	gofmt -w command/plugin.go

test: deps fmt-check ## Run unit tests
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
	@echo "INFO: Packer deps are managed by govendor. See CONTRIBUTING.md"

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: bin checkversion ci default deps fmt fmt-docs fmt-examples generate releasebin test testacc testrace updatedeps
