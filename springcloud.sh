#!/bin/bash

# Config Server
cd support/config-server
./gradlew build
cd ../..
docker build -t someprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/configserver

# Auth server
docker build -t eriklupander/auth-server support/auth-server
docker service rm auth-server
docker service create --replicas 1 -p 9999:9999 --name auth-server --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/auth-server


# Edge Server
cd support/edge-server
./gradlew clean build
cd ../..
docker build -t someprefix/edge-server support/edge-server/
docker service rm edge-server
docker service create --replicas 1 --name edge-server -p 8765:8765 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/edge-server

# Zipkin
docker service rm zipkin
docker service create --constraint node.role==manager --replicas 1 -p 9411:9411 --name zipkin --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 openzipkin/zipkin


# Hystrix Dashboard
docker build -t someprefix/hystrix support/monitor-dashboard
docker service rm hystrix
docker service create --constraint node.role==manager --replicas 1 -p 7979:7979 --name hystrix --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/hystrix

# Turbine
docker service rm turbine
docker service create --constraint node.role==manager --replicas 1 -p 8282:8282 --name turbine --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/turbine
