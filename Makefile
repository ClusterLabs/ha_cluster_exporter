default: clean fmt static-checks test build post-build

download:
	go mod download
	go mod verify

build: ha_cluster_exporter
ha_cluster_exporter: download fmt
	go build .

install:
	go install

static-checks: vet-check fmt-check

vet-check: download
	go vet .

fmt:
	go fmt

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

post-build:
	go mod tidy

release:

.PHONY: default download install static-checks vet-check fmt fmt-check test clean release post-build
