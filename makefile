
OS_EXEC:=
ifeq ($(OS),Windows_NT)
	OS_EXEC =.exe
endif

VERSION_INFO:=github.com/prometheus/common/version

.PHONY: build
build:
	@echo "> building co2_exporter"
	@go build -ldflags "\
		-X ${VERSION_INFO}.Version=`cat VERSION` \
		-X ${VERSION_INFO}.Revision=`git rev-parse HEAD` \
		-X ${VERSION_INFO}.Branch=`git rev-parse --abbrev-ref HEAD` \
		-X ${VERSION_INFO}.BuildUser=${USER} \
		-X ${VERSION_INFO}.BuildDate=`date -u '+%Y-%m-%d_%H:%M:%S'`" \
		-gcflags=-trimpath=$(go env GOPATH) \
		-asmflags=-trimpath=$(go env GOPATH) \
		-o ./bin/co2_exporter${OS_EXEC} \
		main.go
