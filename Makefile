GO    := GO111MODULE=on go
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)
ifeq (arm, $(GOHOSTARCH))
	GOHOSTARM ?= $(shell GOARM= $(GO) env GOARM)
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)v$(GOHOSTARM)
else
	GO_BUILD_PLATFORM ?= $(GOHOSTOS)-$(GOHOSTARCH)
endif
PROMU        := $(FIRST_GOPATH)/bin/promu
PROMU_VERSION ?= 0.13.0
PROMU_URL     := https://github.com/prometheus/promu/releases/download/v$(PROMU_VERSION)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM).tar.gz

# this is the what ends up in the RPM "Version" field and embedded in the --version CLI flag
VERSION ?= $(shell .ci/get_version_from_git.sh)
ifeq ($(VERSION),)
	VERSION := 0.0.0-dev
endif

# if you want to release to OBS, this must be a remotely available Git reference
REVISION ?= $(shell git rev-parse --abbrev-ref HEAD)

# we only use this to comply with RPM changelog conventions at SUSE
AUTHOR ?= shap-staff@suse.de

# you can customize any of the following to build forks
OBS_PROJECT ?= devel:sap:monitoring:factory
REPOSITORY ?= clusterlabs/ha_cluster_exporter

# the Go archs we crosscompile to
ARCHS ?= amd64 arm64 ppc64le s390x

DOCKER_IMAGE_NAME ?= ha_cluster_exporter
DOCKER_IMAGE_TAG  ?= $(VERSION)

default: clean mod-tidy generate fmt vet-check test build

promu-prepare: 
	sed "s/{{.Version}}/$(VERSION)/" .promu.yml >.promu.release.yml
	mkdir -p build/bin

# from https://github.com/prometheus/prometheus/blob/main/Makefile.common
$(PROMU):
	$(eval PROMU_TMP := $(shell mktemp -d))
	curl -s -L $(PROMU_URL) | tar -xvzf - -C $(PROMU_TMP)
	mkdir -p $(FIRST_GOPATH)/bin
	cp $(PROMU_TMP)/promu-$(PROMU_VERSION).$(GO_BUILD_PLATFORM)/promu $(FIRST_GOPATH)/bin/promu
	rm -r $(PROMU_TMP)

build: promu-prepare $(PROMU)
	$(PROMU) build --config .promu.release.yml --prefix=build/bin ha_cluster_exporter-amd64

build-all: clean promu-prepare $(PROMU) $(ARCHS)

$(ARCHS):
	GOOS=linux GOARCH=$@ $(PROMU) build --config .promu.release.yml --prefix=build/bin ha_cluster_exporter-$@

docker: build
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

lint:
	golangci-lint run ./...

install:
	$(GO) install ./cmd/ha_cluster_exporter

static-checks: vet-check fmt-check lint

coverage:
	@mkdir -p build
	$(GO) test -cover -coverprofile=build/coverage ./...
	$(GO) tool cover -html=build/coverage

clean:
	$(GO) clean
	rm -rf build
	rm -f .promu.release.yml

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
	go mod vendor
	tar --sort=name --mtime='UTC 1970-01-01' -c vendor | gzip -n > $@/vendor.tar.gz
	cd $@; osc service manualrun

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
	cd $@; osc service manualrun

dashboards-obs-commit: dashboards-obs-workdir
	cd build/obs/grafana-ha-cluster-dashboards; osc addremove
	cd build/obs/grafana-ha-cluster-dashboards; osc commit -m "Update from git rev $(REVISION)"

.PHONY: $(ARCHS) build build-all checks clean coverage dashboards-obs-commit dashboards-obs-workdir default download \
		exporter-obs-changelog exporter-obs-commit exporter-obs-workdir fmt fmt-check generate install mod-tidy \
		static-checks test vet-check
