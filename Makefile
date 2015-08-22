TEST?=./...
# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)
# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD 2>/dev/null)

default: test dev

ci: deps test

release: updatedeps test bin

bin: deps
	@grep 'const VersionPrerelease = ""' version.go > /dev/null ; if [ $$? -ne 0 ]; then \
		echo "ERROR: You must remove prerelease tags from version.go prior to release."; \
		exit 1; \
	fi
	@sh -c "$(CURDIR)/scripts/build.sh"

deps:
	go get -v -d ./...

dev: deps
	@grep 'const VersionPrerelease = ""' version.go > /dev/null ; if [ $$? -eq 0 ]; then \
		echo "ERROR: You must add prerelease tags to version.go prior to making a dev build."; \
		exit 1; \
	fi
	@PACKER_DEV=1 sh -c "$(CURDIR)/scripts/build.sh"

# generate runs `go generate` to build the dynamically generated
# source files.
generate: deps
	go generate ./...

test: deps
	go test $(TEST) $(TESTARGS) -timeout=15s
	@go vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

# testacc runs acceptance tests
testacc: deps generate
	@echo "WARN: Acceptance tests will take a long time to run and may cost money. Ctrl-C if you want to cancel."
	PACKER_ACC=1 go test -v $(TEST) $(TESTARGS) -timeout=45m

testrace: deps
	go test -race $(TEST) $(TESTARGS) -timeout=15s

# `go get -u` causes git to revert packer to the master branch. This causes all
# kinds of headaches. We record the git sha when make starts try to correct it
# if we detect dift. DO NOT use `git checkout -f` for this because it will wipe
# out your changes without asking.
updatedeps:
	@echo "INFO: Currently on $(GITBRANCH) ($(GITSHA))"
	@git diff-index --quiet HEAD ; if [ $$? -ne 0 ]; then \
		echo "ERROR: Your git working tree has uncommitted changes. updatedeps will fail. Please stash or commit your changes first."; \
		exit 1; \
	fi
	go get -u github.com/mitchellh/gox
	go get -u golang.org/x/tools/cmd/stringer
	go list ./... \
		| xargs go list -f '{{join .Deps "\n"}}' \
		| grep -v github.com/mitchellh/packer \
		| grep -v '/internal/' \
		| sort -u \
		| xargs go get -f -u -v -d ; if [ $$? -eq 0 ]; then \
		echo "ERROR: go get failed. Your git branch may have changed; you were on $(GITBRANCH) ($(GITSHA))."; \
	fi
	@if [ "$(GITBRANCH)" != "" ]; then git checkout -q $(GITBRANCH); else git checkout -q $(GITSHA); fi
	@if [ `git rev-parse HEAD` != "$(GITSHA)" ]; then \
		echo "ERROR: git checkout has drifted and we weren't able to correct it. Was $(GITBRANCH) ($(GITSHA))"; \
		exit 1; \
	fi
	@echo "INFO: Currently on $(GITBRANCH) ($(GITSHA))"

.PHONY: bin checkversion ci default deps generate test testacc testrace updatedeps
