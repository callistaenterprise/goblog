#!/bin/bash

# Config Server
cd support/config-server
./gradlew build
cd ../..
docker build -t someprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/configserver
