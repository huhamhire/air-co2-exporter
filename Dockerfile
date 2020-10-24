FROM golang:1.14.10-alpine3.12 as builder

# Install build dependencies
RUN set -eux; \
    apk add --no-cache git libusb-dev pkgconfig gcc musl-dev make

ENV GOPATH /go/
ENV GO_WORKDIR $GOPATH/src/air-co2-exporter

WORKDIR $GO_WORKDIR

ADD . $GO_WORKDIR

# Fetch Golang Dependency and Build Binary
RUN go install
RUN make build


# Release Alpine Image
FROM alpine:3.12

WORKDIR /usr/local/bin
RUN apk add --no-cache libusb libc6-compat

ENV TZ=UTC LABEL_TAG=default
COPY --from=builder /go/src/air-co2-exporter/bin/air_co2_exporter /usr/local/bin

EXPOSE 9110

ENTRYPOINT ["/usr/local/bin/air_co2_exporter"]
