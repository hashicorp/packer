# Project-specific variables
BINARIES ?=	gotty-client
GOTTY_URL := http://localhost:8081/
VERSION :=	$(shell cat .goxc.json | jq -c .PackageVersion | sed 's/"//g')

CONVEY_PORT ?=	9042


# Common variables
SOURCES :=	$(shell find . -type f -name "*.go")
COMMANDS :=	$(shell go list ./... | grep -v /vendor/ | grep /cmd/)
PACKAGES :=	$(shell go list ./... | grep -v /vendor/ | grep -v /cmd/)
GOENV ?=	GO15VENDOREXPERIMENT=1
GO ?=		$(GOENV) go
GODEP ?=	$(GOENV) godep
USER ?=		$(shell whoami)


all:	build


.PHONY: build
build:	$(BINARIES)


.PHONY: install
install:
	$(GO) install ./cmd/gotty-client


$(BINARIES):	$(SOURCES)
	$(GO) build -i -o $@ ./cmd/$@


.PHONY: test
test:
#	$(GO) get -t .
	$(GO) test -i -v .
	$(GO) test -v .


.PHONY: godep-save
godep-save:
	$(GODEP) save $(PACKAGES) $(COMMANDS)


.PHONY: clean
clean:
	rm -f $(BINARIES)


.PHONY: re
re:	clean all


.PHONY: convey
convey:
	$(GO) get github.com/smartystreets/goconvey
	goconvey -cover -port=$(CONVEY_PORT) -workDir="$(realpath .)" -depth=1


.PHONY:	cover
cover:	profile.out


profile.out:	$(SOURCES)
	rm -f $@
	$(GO) test -covermode=count -coverpkg=. -coverprofile=$@ .


.PHONY: docker-build
docker-build:
	docker build -t moul/gotty-client .
