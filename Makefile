TEST?=./...
export GITSHA?=$(shell git rev-parse HEAD)

default: test vet dev

ci: deps test vet

release: deps test vet bin

# `go get` will sometimes revert to master, which is not what we want in CI.
# We check the git sha when make starts and verify periodically to avoid drift.
# Don't use -f for this because it will wipe out your changes in development.
verifysha:
	@git checkout -q $(GITSHA)
	@if [ `git rev-parse HEAD` != $(GITSHA) ]; then echo "git sha has drifted; aborting"; exit 1; fi

bin: verifysha
	@sh -c "$(CURDIR)/scripts/build.sh"

dev:
	@TF_DEV=1 sh -c "$(CURDIR)/scripts/build.sh"

# generate runs `go generate` to build the dynamically generated
# source files.
generate:
	go generate ./...

test: verifysha
	go test $(TEST) $(TESTARGS) -timeout=10s
	@$(MAKE) vet

# testacc runs acceptance tests
testacc: generate
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package"; \
		exit 1; \
	fi
	PACKER_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 45m

testrace:
	go test -race $(TEST) $(TESTARGS)

updatedeps:
	@echo "Please use `make deps` instead"

deps: verifysha
	go get -u github.com/mitchellh/gox
	go get -u golang.org/x/tools/cmd/stringer
	go list ./... \
		| xargs go list -f '{{join .Deps "\n"}}' \
		| grep -v github.com/mitchellh/packer \
		| grep -v '/internal/' \
		| sort -u \
		| xargs go get -f -u -v -d
	$(MAKE) verifysha

vet: verifysha
	@go vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go vet ./... ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

.PHONY: bin default generate test testacc updatedeps vet deps
