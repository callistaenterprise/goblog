FROM iron/base

ADD swarm-prometheus-discovery-linux-amd64 /
ENTRYPOINT ["./swarm-prometheus-discovery-linux-amd64","-network", "my_network", "-ignoredServices", "prometheus,grafana"]
