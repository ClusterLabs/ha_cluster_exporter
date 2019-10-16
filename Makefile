default: build

build: fmt-check  vet-check test
	go build .
	go mod tidy

install:
	go install

vet-check:
	go vet .

fmt-check:
	go fmt .

test:
	go test -v

coverage:
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: build install vet-check
