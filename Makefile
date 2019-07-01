TEST?=$(shell go list ./...)
VET?=$(shell go list ./...)
# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)
# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
GOPATH=$(shell go env GOPATH)

EXECUTABLE_FILES=$(shell find . -type f -executable | egrep -v '^\./(website/[vendor|tmp]|vendor/|\.git|bin/|scripts/|pkg/)' | egrep -v '.*(\.sh|\.bats|\.git)' | egrep -v './provisioner/(ansible|inspec)/test-fixtures/exit1')

# Get the git commit
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_IMPORT=github.com/hashicorp/packer/version
GOLDFLAGS=-X $(GIT_IMPORT).GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)

export GOLDFLAGS

.PHONY: bin checkversion ci default install-build-deps install-gen-deps fmt fmt-docs fmt-examples generate releasebin test testacc testrace

default: install-build-deps install-gen-deps generate testrace dev releasebin package dev fmt fmt-check mode-check fmt-docs fmt-examples

ci: testrace ## Test in continuous integration

release: install-build-deps test releasebin package ## Build a release build

bin: install-build-deps ## Build debug/test build
	@echo "WARN: 'make bin' is for debug / test builds only. Use 'make release' for release builds."
	@GO111MODULE=auto sh -c "$(CURDIR)/scripts/build.sh"

releasebin: install-build-deps
	@grep 'const VersionPrerelease = "dev"' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must remove prerelease tags from version/version.go prior to release."; \
		exit 1; \
	fi
	@GO111MODULE=auto sh -c "$(CURDIR)/scripts/build.sh"

package:
	$(if $(VERSION),,@echo 'VERSION= needed to release; Use make package skip compilation'; exit 1)
	@sh -c "$(CURDIR)/scripts/dist.sh $(VERSION)"

install-build-deps: ## Install dependencies for bin build
	@go get github.com/mitchellh/gox

install-gen-deps: ## Install dependencies for code generation
	@go get golang.org/x/tools/cmd/goimports
	@./scripts/off_gopath.sh; if [ $$? -eq 0 ]; then \
		go get github.com/mna/pigeon@master; \
	else \
		go get -u github.com/mna/pigeon; \
	fi

	@go get github.com/alvaroloes/enumer

dev: ## Build and install a development build
	@grep 'const VersionPrerelease = ""' version/version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must add prerelease tags to version/version.go prior to making a dev build."; \
		exit 1; \
	fi
	@mkdir -p pkg/$(GOOS)_$(GOARCH)
	@mkdir -p bin
	@go install -ldflags '$(GOLDFLAGS)'
	@cp $(GOPATH)/bin/packer bin/packer
	@cp $(GOPATH)/bin/packer pkg/$(GOOS)_$(GOARCH)

fmt: ## Format Go code
	@go fmt ./...

fmt-check: fmt ## Check go code formatting
	@echo "==> Checking that code complies with go fmt requirements..."
	@git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found files that are not fmt'ed."; \
		echo "You can use the command: \`make fmt\` to reformat code."; \
		exit 1; \
	fi

mode-check: ## Check that only certain files are executable
	@echo "==> Checking that only certain files are executable..."
	@if [ ! -z "$(EXECUTABLE_FILES)" ]; then \
		echo "These files should not be executable or they must be white listed in the Makefile:"; \
		echo "$(EXECUTABLE_FILES)" | xargs -n1; \
		exit 1; \
	else \
		echo "Check passed."; \
	fi
fmt-docs:
	@find ./website/source/docs -name "*.md" -exec pandoc --wrap auto --columns 79 --atx-headers -s -f "markdown_github+yaml_metadata_block" -t "markdown_github+yaml_metadata_block" {} -o {} \;

# Install js-beautify with npm install -g js-beautify
fmt-examples:
	find examples -name *.json | xargs js-beautify -r -s 2 -n -eol "\n"

# generate runs `go generate` to build the dynamically generated
# source files.
generate: install-gen-deps ## Generate dynamically generated code
	go generate ./...
	go fmt common/bootcommand/boot_command.go
	goimports -w common/bootcommand/boot_command.go
	go fmt command/plugin.go

generate-check: generate ## Check go code generation is on par
	@echo "==> Checking that auto-generated code is not changed..."
	@git diff --exit-code; if [ $$? -eq 1 ]; then \
		echo "Found diffs in go generated code."; \
		echo "You can use the command: \`make generate\` to reformat code."; \
		exit 1; \
	fi

test: mode-check vet ## Run unit tests
	@go test $(TEST) $(TESTARGS) -timeout=3m

# testacc runs acceptance tests
testacc: install-build-deps generate ## Run acceptance tests
	@echo "WARN: Acceptance tests will take a long time to run and may cost money. Ctrl-C if you want to cancel."
	PACKER_ACC=1 go test -v $(TEST) $(TESTARGS) -timeout=45m

testrace: mode-check vet ## Test with race detection enabled
	@GO111MODULE=off go test -race $(TEST) $(TESTARGS) -timeout=3m -p=8

check-vendor-vs-mod: ## Check that go modules and vendored code are on par
	@GO111MODULE=on go mod vendor
	@git diff --exit-code --ignore-space-change --ignore-space-at-eol -- vendor ; if [ $$? -eq 1 ]; then \
		echo "ERROR: vendor dir is not on par with go modules definition." && \
		exit 1; \
	fi

vet: ## Vet Go code
	@go vet $(VET)  ; if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
