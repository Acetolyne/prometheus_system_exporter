# Prometheus System Exporter

## Enabling specific stats
Before running the program you should check the settings.ini file located at /etc/prometheus_system_exporter/settings.ini
Enable  any non default stats you would like to have exported to prometheus
Disable any collectors you don't want by specifying the value as false or by commenting out the line with a # at the beginning of the line

## Source settings

By default the source port is 9091, this needs to be specified in the source in your graphing app such as Grafana and can also be updated in the settings file

## Installation
the exporter WILL BE registered as a service but prometheus need to also be running and you need to modify your /etc/prometheus/prometheus.yml file for prometheus to export your custom stats
you should add a target in your jobs section specifying the port number indicated in your /etc/prometheus_system_exporter/settings.ini file, if you have not changed the default port then this will be 9091.
an example of the section in the prometheus.yml file is below.

```
- job_name: node
    # If prometheus-node-exporter is installed, grab stats about the local
    # machine by default.
    static_configs:
        - targets: ['localhost:9100','localhost:9091']
```
