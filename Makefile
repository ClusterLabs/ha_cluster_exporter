# this is the what ends up in the RPM "Version" field and it is also used as suffix for the built binaries
# it can be arbitrary for local builds, but it if you want to commit to OBS it must correspond to a Git tag with an associated GitHub release
VERSION ?= dev

# we only use this to comply with RPM changelog conventions at SUSE
AUTHOR ?= shap-staff@suse.de

# you can customize any of the following to build forks
OBS_PROJECT ?= server:monitoring
OBS_PACKAGE ?= prometheus-ha_cluster_exporter
REPOSITORY ?= clusterlabs/ha_cluster_exporter

# the Go archs we crosscompile to
ARCHS ?= amd64 arm64 ppc64le s390x

default: clean mod-tidy fmt vet-check test build

download:
	go mod download
	go mod verify

build: amd64

build-all: clean-bin $(ARCHS)

$(ARCHS):
	@mkdir -p build/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=$@ go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o build/bin/ha_cluster_exporter-$(VERSION)-$@

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

clean: clean-bin clean-obs
	go clean
	rm -f coverage.out

clean-bin:
	rm -rf build/bin

clean-obs:
	rm -rf build/obs

obs-workdir: clean-obs
	@mkdir -p build/obs/$(OBS_PACKAGE)
	osc checkout $(OBS_PROJECT)/$(OBS_PACKAGE) -o build/obs
	cp ha_cluster_exporter.spec build/obs/$(OBS_PACKAGE).spec
	cp -r doc LICENSE *.md ha_cluster_exporter.service build/obs/$(OBS_PACKAGE)/
	cp build/bin/* build/obs/$(OBS_PACKAGE)/
	mv build/obs/$(OBS_PACKAGE)/ha_cluster_exporter-$(VERSION)-arm64 build/obs/$(OBS_PACKAGE)/ha_cluster_exporter-$(VERSION)-aarch64
	mv build/obs/$(OBS_PACKAGE)/ha_cluster_exporter-$(VERSION)-amd64 build/obs/$(OBS_PACKAGE)/ha_cluster_exporter-$(VERSION)-x86_64
	sed -i 's/%%VERSION%%/$(VERSION)/' build/obs/$(OBS_PACKAGE).spec
	rm build/obs/*.tar.gz
	tar -cvzf build/obs/$(OBS_PACKAGE)-$(VERSION).tar.gz -C build/obs/$(OBS_PACKAGE) .
	.ci/gh_release_to_obs_changeset.py $(REPOSITORY) --author $(AUTHOR) --tag $(VERSION) -f build/obs/$(OBS_PACKAGE).changes

obs-commit: obs-workdir
	cd build/obs; osc addremove
	cd build/obs; osc commit -m "Automated $(VERSION) release"

.PHONY: default download install static-checks vet-check fmt fmt-check mod-tidy test clean clean-bin clean-obs build build-all obs-commit obs-workdir $(ARCHS)
