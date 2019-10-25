default: clean check test build

build: ha_cluster_exporter
ha_cluster_exporter:
	go fmt
	go build .

install:
	go install

check: vet-check fmt-check

vet-check:
	go vet .

fmt-check:
	.ci/go_lint.sh

test:
	go test -v

coverage: coverage.out
coverage.out:
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	go clean
	rm -f coverage.out

.PHONY: default install check vet-check fmt-check test clean
