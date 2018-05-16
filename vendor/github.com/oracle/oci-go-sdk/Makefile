DOC_SERVER_URL=https:\/\/docs.us-phoenix-1.oraclecloud.com

GEN_TARGETS = identity core objectstorage loadbalancer database audit
NON_GEN_TARGETS = common common/auth
TARGETS = $(NON_GEN_TARGETS) $(GEN_TARGETS)

TARGETS_WITH_TESTS = common common/auth
TARGETS_BUILD = $(patsubst %,build-%, $(TARGETS))
TARGETS_CLEAN = $(patsubst %,clean-%, $(GEN_TARGETS))
TARGETS_LINT = $(patsubst %,lint-%, $(TARGETS))
TARGETS_TEST = $(patsubst %,test-%, $(TARGETS_WITH_TESTS))
TARGETS_RELEASE= $(patsubst %,release-%, $(TARGETS))
GOLINT=$(GOPATH)/bin/golint
LINT_FLAGS=-min_confidence 0.9 -set_exit_status


.PHONY: $(TARGETS_BUILD) $(TARGET_TEST)

build: lint $(TARGETS_BUILD)

test: build $(TARGETS_TEST)

lint: $(TARGETS_LINT)

clean: $(TARGETS_CLEAN)

$(TARGETS_LINT): lint-%:%
	@echo "linting and formatting: $<"
	@(cd $< && gofmt -s -w .)
	@if [ \( $< = common \) -o \( $< = common/auth \) ]; then\
		(cd $< && $(GOLINT) -set_exit_status .);\
	else\
		(cd $< && $(GOLINT) $(LINT_FLAGS) .);\
	fi

$(TARGETS_BUILD): build-%:%
	@echo "building: $<"
	@(cd $< && find . -name '*_integ_test.go' | xargs -I{} mv {} ../integtest)
	@(cd $< && go build -v)

$(TARGETS_TEST): test-%:%
	@(cd $< && go test -v)

$(TARGETS_CLEAN): clean-%:%
	@echo "cleaning $<"
	@-rm -rf $<

pre-doc:
	@echo "Rendering doc server to ${DOC_SERVER_URL}"
	find . -name \*.go |xargs sed -i '' 's/{{DOC_SERVER_URL}}/${DOC_SERVER_URL}/g'

gen-version:
	go generate -x

release: gen-version build pre-doc
