#!/bin/bash

docker service rm prometheus
docker service create -p 9090:9090 --constraint node.role==manager --mount type=volume,source=swarm-endpoints,target=/etc/swarm-endpoints/,volume-driver=local --name=prometheus --replicas=1 --network=my_network someprefix/prometheus
