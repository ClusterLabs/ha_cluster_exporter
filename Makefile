default: build

build: fmt-check  vet-check 
	go build .

install:
	go install

vet-check:
	go vet .

fmt-check:
	go fmt  .

# This deploy the binary to a node of cluster in devel mode port :9002. 
# you need to change the IP adress in the script. 
# TODO: (In future we can add an arg var..)
deploy:
	tools/deploy-to-cluster.sh
.PHONY: build install vet-check
