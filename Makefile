default: build

build: fmt-check  vet-check test
	go build .

install:
	go install

vet-check:
	go vet .

fmt-check:
	go fmt  .
test:
	go test 

coverage:
	go test -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out
# This deploy the binary to a node of cluster in devel mode port :9002. 
# you need to change the IP adress in the script. 
# TODO: (In future we can add an arg var..)
deploy:
	tools/deploy-to-cluster.sh
.PHONY: build install vet-check
