# Local dev helper. The authoritative release build lives in
# .github/workflows/release.yml. version.txt is the single source of truth;
# `make build` injects it the same way CI does. A plain `go build` (no ldflags)
# leaves VERSION as "dev" by design.

VERSION := $(shell tr -d '[:space:]' < version.txt)
LDFLAGS := -s -w -X main.VERSION=$(VERSION)

.PHONY: build test vet fmt-check clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o basicToOauth .

test:
	go test -race ./...

vet:
	go vet ./...

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then echo "gofmt needed on:"; echo "$$out"; exit 1; fi

clean:
	rm -rf build basicToOauth basicToOauth.exe
