NAME = couchlock
BUILDDIR = ./ARTIFACTS

# Remove prefix since deb, rpm etc don't recognize this as valid version
VERSION = $(shell git describe --tags --match 'v[0-9]*\.[0-9]*\.[0-9]*' | sed 's/^v//')

###############################################################################
## Building
###############################################################################

.PHONY: build build_darwin build_linux
build: build_darwin build_linux

compile = bash -c "env GOOS=$(1) GOARCH=$(2) go build -a \
						-ldflags \"-w -X main.Version='$(VERSION)'\" \
						-o $(BUILDDIR)/$(NAME)-$(VERSION)-$(1)-$(2)"

build_darwin:
	$(call compile,darwin,amd64)

build_linux:
	$(call compile,linux,amd64)


###############################################################################
## Clean
##
## EXPLICITLY removing artifacts directory to protect from horrible accidents
###############################################################################
clean:
	rm -rf ./ARTIFACTS

###############################################################################
## Test using bats
##
## bats and docker need to be installed and go binary path must be in your PATH
###############################################################################
test:
	go build && go install && bats ./bats
