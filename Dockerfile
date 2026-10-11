FROM golang:1.27-alpine3.24 AS builder

# Install build dependencies
RUN apk add --no-cache git libusb-dev pkgconf gcc musl-dev make
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 make build


# Release Alpine Image
FROM alpine:3.24

WORKDIR /usr/local/bin
RUN apk add --no-cache libusb

ENV TZ=UTC LABEL_TAG=default
COPY --from=builder /src/bin/air_co2_exporter /usr/local/bin

EXPOSE 9110

ENTRYPOINT ["/usr/local/bin/air_co2_exporter"]
