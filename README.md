# Air CO2 Exporter

Prometheus metrics exporter for TFA Dostmann air CO2 monitor.

## Usage


### Systemd

`/etc/systemd/system/air_co2_exporter.service`

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

```bash
systemctl daemon-reload

systemctl start air_co2_exporter
systemctl enable air_co2_exporter
```

### Docker

### Kubernetes
