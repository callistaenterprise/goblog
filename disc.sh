#!/bin/bash
export GOOS=linux
export CGO_ENABLED=0

cd swarm-prometheus-discovery;go get;go build -o swarm-prometheus-discovery-linux-amd64;echo built `pwd`;cd ..

export GOOS=darwin

docker build -t someprefix/swarm-prometheus-discovery swarm-prometheus-discovery/
docker service rm swarm-prometheus-discovery
docker service create  --constraint node.role==manager --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock --name=swarm-prometheus-discovery --replicas=1 --network=my_network someprefix/swarm-prometheus-discovery
