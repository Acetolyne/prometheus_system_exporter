# Prometheus System Exporter

## Enabling specific stats
Before running the program you should check the settings.ini file located at /etc/prometheus_system_exporter/settings.ini
Enable  any non default stats you would like to have exported to prometheus
Disable any collectors you don't want by specifying the value as false or by commenting out the line with a # at the beginning of the line

## Source settings

By default the source port is 9091, this needs to be specified in the source in your graphing app such as Grafana and can also be updated in the settings file