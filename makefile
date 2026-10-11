VERSION = $(shell cat VERSION)

INFO_PATH := github.com/prometheus/common/version
APP_NAME = air-co2-exporter
BIN_NAME = air_co2_exporter
DOCKER_REGISTRY ?=

OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
DOCKER_REPOSITORY = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/)$(APP_NAME)
DOCKER_IMAGE = $(DOCKER_REPOSITORY):$(VERSION)-$(OS)-$(ARCH)
REVISION := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)
BUILD_USER := $(shell whoami)
BUILD_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')

.PHONY: lint
lint:
	@test -z "$$(gofmt -l main.go monitor prometheus)"
	go vet ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: build
build:
	@echo "> building ${BIN_NAME}"
	@CGO_ENABLED=1 GOOS=$(OS) GOARCH=$(ARCH) go build -trimpath -ldflags "\
		-X ${INFO_PATH}.Version=${VERSION} \
		-X ${INFO_PATH}.Revision=${REVISION} \
		-X ${INFO_PATH}.Branch=${BRANCH} \
		-X ${INFO_PATH}.BuildUser=${BUILD_USER} \
		-X ${INFO_PATH}.BuildDate=${BUILD_DATE}" \
		-o ./bin/${BIN_NAME} \
		.

.PHONY: archive docker-build docker-push docker-run docker-merge-manifest
archive:
	@mkdir -p ./dist
	tar -czf ./dist/${BIN_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz \
		bin/${BIN_NAME} \
		README.md \
		LICENSE

docker-build:
	docker build -f Dockerfile \
		-t $(DOCKER_IMAGE) \
		.

docker-push:
	docker push $(DOCKER_IMAGE)

docker-run:
	docker run --privileged -p 9110:9110 $(DOCKER_IMAGE)

docker-merge-manifest:
	docker manifest create \
		$(DOCKER_REPOSITORY):latest \
		$(DOCKER_REPOSITORY):$(VERSION)-linux-amd64 \
		$(DOCKER_REPOSITORY):$(VERSION)-linux-arm64
	docker manifest create \
		$(DOCKER_REPOSITORY):$(VERSION) \
		$(DOCKER_REPOSITORY):$(VERSION)-linux-amd64 \
		$(DOCKER_REPOSITORY):$(VERSION)-linux-arm64
	docker manifest push --purge $(DOCKER_REPOSITORY):latest
	docker manifest push --purge $(DOCKER_REPOSITORY):$(VERSION)

.PHONY: clean
clean:
	rm -rf ./bin/* ./dist/*
