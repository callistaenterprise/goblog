#!/bin/bash
docker service rm dvizz
docker service create --constraint node.role==manager --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock --name=dvizz --replicas=1 --network=my_network -p=6969:6969 eriklupander/dvizz
