# Air CO2 Exporter

[![Build Status](https://travis-ci.org/huhamhire/air-co2-exporter.svg?branch=master)](https://travis-ci.org/huhamhire/air-co2-exporter)
[![Docker Image Pulls](https://img.shields.io/docker/pulls/huhamhire/air-co2-exporter.svg)](https://img.shields.io/docker/pulls/huhamhire/air-co2-exporter.svg)

Prometheus exporter for TFA Dostmann air CO2 monitor.

Air CO2 Exporter could read CO2 concentration and indoor temperature metrics from an AirCO2NTROL device.


This project is a golang implementation ported from the old Node.js project which shares similar functions. The new exporter reduced deployment file size and runtime dependencies, added support for docker and kubernetes enviroment, and more easy to use.

Node.js exporter: [https://github.com/huhamhire/co2-monitor-exporter](https://github.com/huhamhire/co2-monitor-exporter)

## Contents

* [Installation](#installation)
  * [Requirements](#requirements)
  * [Dependencies](#dependencies)
  * [Install](#install)
* [Usage](#usage)
  * [CLI](#cli)
  * [Systemd](#systemd)
  * [Docker](#docker)
  * [Kubernetes](#kubernetes)
* [Metrics](#metrics)
* [Compatible Devices](#compatible-devices)
* [Troubleshooting](#troubleshooting)
* [License](#license)

## Installation

### Requirements

* Operating System: Linux
* CPU Architecture Supported:
  - amd64
  - arm64

Excutable files can be built for other operating systems like Windows, OS X, or other platforms like ARM devices. However, there could be issues related to libusb library on OS other than Linux.

### Dependencies

`air_co2_exporter` used `libusb` to read data from AirCO2NTROL devices.

```bash
# Debian / Ubuntu
sudo apt-get install -y libusb-1.0

# RHEL / CentOS
sudo yum install -y libusb-1.0

# Arch
sudo pacman -S libusb

# Alpine Linux
apk add libusb
```

### Install

Download the `air_co2_exporter` excutable file and put it under system path like `/usr/local/bin/`.


## Usage

### CLI

`air_co2_exporter` can be used as a normal program in foreground by shell.

```bash
$ air_co2_exporter --help
usage: air_co2_exporter [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9110"
                             Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"
                             Path under which to expose metrics.
  -t, --label.tag="default"  Tag for exposed metrics.
      --log.level=info       Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt, json]
      --version              Show application version.
```

### Systemd

If you perferred to manage `air_co2_exporter` by systemd, a systemd unit file should be created.

* `/etc/systemd/system/air_co2_exporter.service`

  ```unit file (systemd)
  [Unit]
  Description=Air Co2 Exporter
  
  [Service]
  User=prometheus
  Group=prometheus
  Restart=on-failure
  ExecStart=/usr/local/bin/air_co2_exporter \
          --web.listen-address=:9110 \
          --log.level=warn
  
  [Install]
  WantedBy=default.target
  ```

Once unit file is created, you can manage `air_co2_exporter` like other systemd services.

```bash
systemctl daemon-reload

systemctl start air_co2_exporter
systemctl enable air_co2_exporter
```

### Docker

`air_co2_exporter` is supported to be run in a docker container environment with privileged mode. 

[https://hub.docker.com/r/huhamhire/air-co2-exporter](https://hub.docker.com/r/huhamhire/air-co2-exporter)

A custom metrics tag can be set by environment virable `LABEL_TAG`.

```bash
docker run --privileged -p 9110:9110 huhamhire/air-co2-exporter:latest
```

### Kubernetes

In addition to run with docker, `air-co2-exporter` image is supported to be managed in a kubernetens cluster. 

Node(s) attached with `AirCO2NTROL` device can be managed by a specific label. And exporter pod could be deployed to these nodes later.

```bash
kubectl label node ${node_name} alpha.monitor.device/air-co2=exists
```

However, without a kubernetes device plugin implementation, **only** one `air-co2-exporter` pod can be managed on a single node. Use this exporter as `DaemonSet` is recommended in most scenarios.

**⚠ IMPORTANT**: Pod `securityContext` must be set to privileged mode for the exporter to access usb devices on kubernetes host node. Otherwise, it will enconter program panic cause by `libusb error -99`.

Here is an example kubernetes DaemonSet configuration.

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: air-co2-exporter
  namespace: temperature
  labels:
    app: air-co2-exporter
    prometheus: exporter
    sensor: air-co2
spec:
  selector:
    matchLabels:
      app: air-co2-exporter
  updateStrategy:
    type: RollingUpdate
  revisionHistoryLimit: 1
  template:
    metadata:
      labels:
        app: air-co2-exporter
      annotations:
        prometheus.io/port: '9110'
        prometheus.io/scrape: 'true'
    spec:
      restartPolicy: Always
      terminationGracePeriodSeconds: 0
      nodeSelector:
        alpha.monitor.device/air-co2: exists
      containers:
        - name: exporter
          image: huhamhire/air-co2-exporter:latest
          imagePullPolicy: IfNotPresent
          ports:
            - name: metrics
              hostPort: 9110
              containerPort: 9110
              protocol: TCP
          env:
            - name: LABEL_TAG
              value: my_custom_tag
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
          securityContext:
            privileged: true
            capabilities:
              add:
                - SYS_ADMIN
              drop:
                - ALL
            readOnlyRootFilesystem: true
            allowPrivilegeEscalation: true
```


## Metrics

* `air_temp` - Ambient Temperature (Tamb) in ℃.
* `air_co2` - Relative Concentration of CO2 (CntR) in ppm. 


## Compatible Devices

* [TFA Dostmann AirCO2NTROL Mini - Monitor CO2 31.5006.02](https://www.amazon.de/dp/B00TH3OW4Q)

Other `AirCO2NTROL` devices with more metrics has not been tested. Supplement for compatible device list is always welcomed.


## Limitations

* Only support Linux amd64 and arm64 platforms currently.
* `air_co2_exporter` could read metrics from only one `AirCO2NTROL` device only.
* Device ID (`VID: 0x04D9, PID: 0xA052`) of AirCO2NTROL Mini is hard coded in the exporterd currently.


## Troubleshooting

* libusb error -99

   ```
   panic: libusb: unknown error [code -99]
   ```

   This could happen when `air_co2_exporter` does not have sufficient previleges to access usb devices. Mostly caused by using a docker container environment without `privileged` mode.

* `libusb` not found
 
   ```
   air_co2_exporter: error while loading shared libraries: libusb-1.0.so.0: cannot open shared object file: No such file or directory
   ```
   
   `libusb` is required by `gousb` which used by `air-co2-exporter` to communicate with the device. `libusb-1.0` library is required to be installed first.
  
* `air_co2_exporter` is not found by shell.

   ```
   sh: air_co2_exporter: not found
   ```
  
   Typically seen when used in Apline Linux docker images. Since Alpine image is a compact docker image with the size of only around 5MB. It does not come with a GNU C library (glibc) by default which golang programs required. A `glibc` library like `libc6-compat` on Alpine is required.
  

## License

[MIT](./LICENSE)


---
Made on 🌍 with 💓.
