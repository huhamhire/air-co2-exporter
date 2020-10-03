VERSION = $(shell cat VERSION)

OS_EXEC:=
ifeq ($(OS),Windows_NT)
	OS_EXEC =.exe
endif

INFO_PATH:=github.com/prometheus/common/version
DOCKER_IMG=co2-exporter

.PHONY: build
build:
	@echo "> building co2_exporter"
	@go build -ldflags "\
		-X ${INFO_PATH}.Version=${VERSION} \
		-X ${INFO_PATH}.Revision=`git rev-parse HEAD` \
		-X ${INFO_PATH}.Branch=`git rev-parse --abbrev-ref HEAD` \
		-X ${INFO_PATH}.BuildUser=${USER} \
		-X ${INFO_PATH}.BuildDate=`date -u '+%Y-%m-%d_%H:%M:%S'`" \
		-gcflags "all=-trimpath=${GOPATH}" \
		-o ./bin/co2_exporter${OS_EXEC} \
		main.go

docker-build:
	docker build -f Dockerfile -t ${DOCKER_IMG}:${VERSION} .

docker-run:
	docker run --privileged ${DOCKER_IMG}:${VERSION}

.PHONY: clean
clean:
	rm -rf co2_exporter
