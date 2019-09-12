default: build

build: fmt-check  vet-check 
	go build .

install:
	go install

vet-check:
	go vet .

fmt-check:
	go fmt  .

.PHONY: build install vet-check
