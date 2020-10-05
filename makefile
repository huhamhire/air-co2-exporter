VERSION = $(shell cat VERSION)

INFO_PATH := github.com/prometheus/common/version
APP_NAME = air-co2-exporter
BIN_NAME = air_co2_exporter
DOCKER_REGISTRY ?=

GOPATH ?= `pwd`/../../
OS ?= linux
ARCH ?= amd64

.PHONY: build
build:
	@echo "> building ${BIN_NAME}"
	@GOPATH=${GOPATH} go build -ldflags "\
		-X ${INFO_PATH}.Version=${VERSION} \
		-X ${INFO_PATH}.Revision=`git rev-parse HEAD` \
		-X ${INFO_PATH}.Branch=`git rev-parse --abbrev-ref HEAD` \
		-X ${INFO_PATH}.BuildUser=${USER} \
		-X ${INFO_PATH}.BuildDate=`date -u '+%Y-%m-%d_%H:%M:%S_UTC'`" \
		-gcflags "all=-trimpath=${GOPATH}" \
		-o ./bin/${BIN_NAME} \
		main.go

archive:
	@mkdir -p ./dist
	tar -czf ./dist/${BIN_NAME}-${VERSION}-${OS}-${ARCH}.tar.gz \
		bin/${BIN_NAME} \
		LICENSE

docker-build:
	docker build -f Dockerfile -t ${APP_NAME}:${VERSION} .

docker-run:
	docker run --privileged ${APP_NAME}:${VERSION}

docker-push:
	docker tag ${APP_NAME}:${VERSION} ${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}
	docker tag ${APP_NAME}:${VERSION} ${DOCKER_REGISTRY}/${APP_NAME}:latest
	docker push ${DOCKER_REGISTRY}/${APP_NAME}:${VERSION}
	docker push ${DOCKER_REGISTRY}/${APP_NAME}:latest

.PHONY: clean
clean:
	rm -rf ./bin/* ./dist/*
