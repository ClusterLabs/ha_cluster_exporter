# this is the what ends up in the RPM "Version" field and it is also used as suffix for the built binaries
# if you want to release to OBS it must be a remotely available Git reference
VERSION ?= $(shell git describe --tags --abbrev=0)dev+git.$(shell git show -s --format=%ct.%h HEAD)
DATE = $(shell date --iso-8601=seconds)

# we only use this to comply with RPM changelog conventions at SUSE
AUTHOR ?= shap-staff@suse.de

# you can customize any of the following to build forks
OBS_PROJECT ?= server:monitoring
OBS_PACKAGE ?= prometheus-ha_cluster_exporter
REPOSITORY ?= clusterlabs/ha_cluster_exporter

# the Go archs we crosscompile to
ARCHS ?= amd64 arm64 ppc64le s390x

default: clean download mod-tidy fmt vet-check test build

download:
	go mod download
	go mod verify

build: amd64

build-all: clean-bin $(ARCHS)

$(ARCHS):
	@mkdir -p build/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=$@ go build -trimpath -ldflags "-s -w -X main.version=$(VERSION) -X main.buildDate=$(DATE)" -o build/bin/ha_cluster_exporter-$(VERSION)-$@

install:
	go install

static-checks: vet-check fmt-check

vet-check: download
	go vet ./...

fmt:
	go fmt ./...

mod-tidy:
	go mod tidy

fmt-check:
	.ci/go_lint.sh

test: download
	go test -v ./...

coverage:
	@mkdir -p build
	go test -cover -coverprofile=build/coverage ./...
	go tool cover -html=build/coverage

clean:
	go clean
	rm -rf build

prepare-exporter-release:
	rm -rf build/obs/exporter
	@mkdir -p build/obs/exporter
	osc checkout $(OBS_PROJECT) $(OBS_PACKAGE) -o build/obs/exporter
	rm -f build/obs/exporter/*.tar.gz
	cp -rv packaging/obs/prometheus-ha_cluster_exporter/* build/obs/exporter/
# we interpolate environment variables in OBS _service file so that we control what is downloaded by the tar_scm source service
	sed -i 's~%%VERSION%%~$(VERSION)~' build/obs/exporter/_service
	sed -i 's~%%REPOSITORY%%~$(REPOSITORY)~' build/obs/exporter/_service
	cd build/obs; osc service runall
	.ci/gh_release_to_obs_changeset.py $(REPOSITORY) -a $(AUTHOR) -t $(VERSION) -f build/obs/exporter/$(OBS_PACKAGE).changes || true

exporter-release: prepare-exporter-release
	cd build/obs/exporter; osc addremove
	cd build/obs/exporter; osc commit -m "Update to version $(VERSION)"

.PHONY: default download install static-checks vet-check fmt fmt-check mod-tidy test coverage clean clean-bin clean-obs build build-all obs-commit obs-workdir $(ARCHS)
