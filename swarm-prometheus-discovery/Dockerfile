FROM alpine:latest

ADD bin/swarm-prometheus-discovery /
ENTRYPOINT ["./swarm-prometheus-discovery","-network", "my_network", "-ignoredServices", "goblog_prometheus,goblog_grafana"]
