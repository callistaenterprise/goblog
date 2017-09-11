#!/bin/bash

# Config Server
cd support/config-server
./gradlew build
cd ../..
docker build -t someprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/configserver

# Hystrix Dashboard
docker build -t someprefix/hystrix support/monitor-dashboard
docker service rm hystrix
docker service create --constraint node.role==manager --replicas 1 -p 7979:7979 --name hystrix --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/hystrix

# Turbine
docker service rm turbine
docker service create --constraint node.role==manager --replicas 1 -p 8282:8282 --name turbine --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/turbine
