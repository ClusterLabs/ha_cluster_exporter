# this is the what ends up in the RPM "Version" field and embedded in the --version CLI flag
VERSION ?= $(shell .ci/get_version_from_git.sh)

# this will be used as the build date by the Go compile task
DATE = $(shell date --iso-8601=seconds)

# if you want to release to OBS, this must be a remotely available Git reference
REVISION ?= $(shell git rev-parse --abbrev-ref HEAD)

# we only use this to comply with RPM changelog conventions at SUSE
AUTHOR ?= shap-staff@suse.de

# you can customize any of the following to build forks
OBS_PROJECT ?= network:ha-clustering:sap-deployments:devel
REPOSITORY ?= clusterlabs/ha_cluster_exporter

# the Go archs we crosscompile to
ARCHS ?= amd64 arm64 ppc64le s390x

default: clean mod-tidy generate fmt vet-check test build

build: amd64

build-all: clean $(ARCHS)

$(ARCHS):
	@mkdir -p build/bin
	CGO_ENABLED=0 GOOS=linux GOARCH=$@ go build -trimpath -ldflags "-s -w -X main.version=$(VERSION) -X main.buildDate=$(DATE)" -o build/bin/ha_cluster_exporter-$@

install:
	go install

static-checks: vet-check fmt-check

vet-check:
	go vet ./...

fmt:
	go fmt ./...

mod-tidy:
	go mod tidy

fmt-check:
	.ci/go_lint.sh

generate:
	go generate ./...

test:
	go test -v ./...

checks: static-checks test

coverage:
	@mkdir -p build
	go test -cover -coverprofile=build/coverage ./...
	go tool cover -html=build/coverage

clean:
	go clean
	rm -rf build

exporter-obs-workdir: build/obs/prometheus-ha_cluster_exporter
build/obs/prometheus-ha_cluster_exporter:
	@mkdir -p $@
	osc checkout $(OBS_PROJECT) prometheus-ha_cluster_exporter -o $@
	rm -f $@/*.tar.gz
	cp -rv packaging/obs/prometheus-ha_cluster_exporter/* $@/
# we interpolate environment variables in OBS _service file so that we control what is downloaded by the tar_scm source service
	sed -i 's~%%VERSION%%~$(VERSION)~' $@/_service
	sed -i 's~%%REVISION%%~$(REVISION)~' $@/_service
	sed -i 's~%%REPOSITORY%%~$(REPOSITORY)~' $@/_service
	cd $@; osc service runall

exporter-obs-changelog: exporter-obs-workdir
	.ci/gh_release_to_obs_changeset.py $(REPOSITORY) -a $(AUTHOR) -t $(REVISION) -f build/obs/prometheus-ha_cluster_exporter/prometheus-ha_cluster_exporter.changes

exporter-obs-commit: exporter-obs-workdir
	cd build/obs/prometheus-ha_cluster_exporter; osc addremove
	cd build/obs/prometheus-ha_cluster_exporter; osc commit -m "Update from git rev $(REVISION)"

dashboards-obs-workdir: build/obs/grafana-ha-cluster-dashboards
build/obs/grafana-ha-cluster-dashboards:
	@mkdir -p $@
	osc checkout $(OBS_PROJECT) grafana-ha-cluster-dashboards -o $@
	rm -f $@/*.tar.gz
	cp -rv packaging/obs/grafana-ha-cluster-dashboards/* $@/
# we interpolate environment variables in OBS _service file so that we control what is downloaded by the tar_scm source service
	sed -i 's~%%REVISION%%~$(REVISION)~' $@/_service
	sed -i 's~%%REPOSITORY%%~$(REPOSITORY)~' $@/_service
	cd $@; osc service runall

dashboards-obs-commit: dashboards-obs-workdir
	cd build/obs/grafana-ha-cluster-dashboards; osc addremove
	cd build/obs/grafana-ha-cluster-dashboards; osc commit -m "Update from git rev $(REVISION)"

.PHONY: $(ARCHS) build build-all checks clean coverage dashboards-obs-commit dashboards-obs-workdir default download \
		exporter-obs-changelog exporter-obs-commit exporter-obs-workdir fmt fmt-check generate install mod-tidy \
		static-checks test vet-check
