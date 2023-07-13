COMMIT := $(shell git describe --dirty --long --always)
VERSION := $(shell cat ./VERSION)
VERSION := $(VERSION)-$(COMMIT)
ARCH := $(shell dpkg --print-architecture)

default: build ;

prepare:
	@go mod tidy
	@mkdir -p dist

build: prepare
	GOOS=linux CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/bima_${ARCH} ./cmd/bima

install:
	mv ./dist/bima_${ARCH} /usr/local/bin/bima

uninstall:
	rm -f /usr/local/bin/bima

clean:
	@rm -fr ./dist/

build_aarch64: prepare
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/bima_aarch64 ./cmd/bima

build_amd64: prepare
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -ldflags "-w" -ldflags "-linkmode 'external' -extldflags '-static'" \
          -ldflags "-X main.version=${VERSION}" -o ./dist/bima_amd64 ./cmd/bima

all: build_aarch64 build_amd64
