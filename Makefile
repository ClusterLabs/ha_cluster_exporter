VERSION ?= dev
ARCHS = amd64 arm64 ppc64le s390x

default: clean mod-tidy fmt vet-check test build

download:
	go mod download
	go mod verify

build: amd64

build-all: clean $(ARCHS)

$(ARCHS):
	@mkdir -p build
	CGO_ENABLED=0 GOOS=linux GOARCH=$@ go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o build/ha-cluster-exporter-${VERSION}-linux-$@

install:
	go install

static-checks: vet-check fmt-check

vet-check: download
	go vet .

fmt:
	go fmt

mod-tidy:
	go mod tidy

fmt-check:
	.ci/go_lint.sh

test: download
	go test -v

coverage: coverage.out
coverage.out:
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	go clean
	rm -f coverage.out
	rm -rf build/*

obs-commit:

.PHONY: default download install static-checks vet-check fmt fmt-check mod-tidy test clean build build-all obs-commit $(ARCHS)
