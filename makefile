
build:
	go build -ldflags "\
		-X main.buildstamp `date -u '+%Y-%m-%d_%I:%M:%S%p'` \
		-X main.githash `git rev-parse HEAD`" \
		-o ./bin/monitor.exe \
		main.go

.PHONY: build
