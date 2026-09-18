# Prometheus System Exporter

## Permissions
The exporter needs to be run as a root user or admin on windows, this allows the collector the gather the required information to report
metrics.

## Service
You typically want to set the exporter up to run as a service so you do not need to start it manually each time the system starts.

### Linux
to add the exporter as a service, create the file /etc/systemd/system/prometheus-system-exporter.service and add the following content to the file, then save it.

```ini
[Unit]
Description=System Prometheus Exporter
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/sbin/prometheus-system-exporter
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Then enable and start it:

```
sudo systemctl daemon-reload
sudo systemctl enable --now prometheus-system-exporter
```

### Windows
Add this later

## Enabling specific stats
Before running the program you should check the settings.ini file located at /etc/prometheus-system-exporter/settings.ini
Enable  any non default stats you would like to have exported to prometheus
Disable any collectors you don't want by specifying the value as false or by commenting out the line with a # at the beginning of the line

On Linux, if this file doesn't exist yet, the exporter will create the directory and a default settings.ini for you on first run.

## Source settings

By default the source port is 9091, this needs to be specified in the source in your graphing app such as Grafana and can also be updated in the settings file

## Installation
the exporter WILL BE registered as a service but prometheus need to also be running and you need to modify your /etc/prometheus/prometheus.yml file for prometheus to export your custom stats
you should add a target in your jobs section specifying the port number indicated in your /etc/prometheus-system-exporter/settings.ini file, if you have not changed the default port then this will be 9091.
an example of the section in the prometheus.yml file is below.

  - job_name: node
    # If prometheus-node-exporter is installed, grab stats about the local
    # machine by default.
    static_configs:
            - targets: ['localhost:9100','localhost:9091']
