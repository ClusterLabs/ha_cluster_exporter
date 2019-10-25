default: clean static-checks test build

download:
	go mod download
	go mod verify

build: ha_cluster_exporter
ha_cluster_exporter: download
	go fmt
	go build .

install:
	go install

static-checks: vet-check fmt-check

vet-check: download
	go vet .

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

release:

.PHONY: default download install static-checks vet-check fmt-check test clean release
