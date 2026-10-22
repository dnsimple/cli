UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  SED := $(shell command -v gsed 2>/dev/null)
else
  SED := sed
endif

BINARY_NAME := dnsimple
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build install test lint clean nix-update-hash

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/dnsimple

install:
	go install $(LDFLAGS) ./cmd/dnsimple

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf bin/

nix-update-hash:
ifndef SED
	$(error gsed is required on macOS — install with: brew install gnu-sed)
endif
	@OLD_HASH=$$(grep 'vendorHash' default.nix | $(SED) 's/.*"\(.*\)".*/\1/') && \
		$(SED) -i 's|vendorHash = ".*"|vendorHash = lib.fakeHash|' default.nix && \
		NEW_HASH=$$(nix-build -E 'with import <nixpkgs> {}; callPackage ./default.nix {}' 2>&1 | grep 'got:' | awk '{print $$2}') && \
		$(SED) -i "s|vendorHash = lib.fakeHash|vendorHash = \"$$NEW_HASH\"|" default.nix && \
		echo "Updated vendorHash to $$NEW_HASH" || \
		($(SED) -i "s|vendorHash = lib.fakeHash|vendorHash = \"$$OLD_HASH\"|" default.nix && echo "Failed to compute hash" && exit 1)
