FROM alpine:3.12

WORKDIR /usr/local/bin
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' \
    /etc/apk/repositories && \
    apk add --no-cache libusb libc6-compat

ENV TZ=UTC LABEL_TAG=default
ADD [ "./bin/co2_exporter", "/usr/local/bin" ]

EXPOSE 9110

ENTRYPOINT ["/usr/local/bin/co2_exporter"]
