#!/bin/bash

# Config Server
cd support/config-server
./gradlew build
cd ../..
docker build -t someprefix/configserver support/config-server/
docker service rm configserver
docker service create --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/configserver

# Auth Server
cd support/auth-server
./gradlew clean build
cd ../..
docker build -t someprefix/auth-server support/auth-server/
docker service rm auth-server
docker service create --replicas 1 --name auth-server -p 9999:9999 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/auth-server



# Edge Server
cd support/edge-server
./gradlew clean build
cd ../..
docker build -t someprefix/edge-server support/edge-server/
docker service rm edge-server
docker service create --replicas 1 --name edge-server -p 8765:8765 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/edge-server

