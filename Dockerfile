FROM alpine:3.12

WORKDIR /usr/local/bin
RUN apk add --no-cache libusb libc6-compat

ENV TZ=UTC LABEL_TAG=default
ADD [ "./bin/air_co2_exporter", "/usr/local/bin" ]

EXPOSE 9110

ENTRYPOINT ["/usr/local/bin/air_co2_exporter"]
