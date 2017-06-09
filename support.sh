#!/bin/bash

# RabbitMQ
docker service rm rabbitmq
docker build -t someprefix/rabbitmq support/rabbitmq/
docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 someprefix/rabbitmq

# CoachroachDB
docker service rm cockroachdb1
docker service create --name=cockroachdb1 --network=my_network -p 26257:26257 -p 3030:8080 --mount type=volume,source=cockroach-data1,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.0.1 start --insecure

# CoachroachDB  -p 26258:26257 -p 3031:8080 
docker service rm cockroachdb2
docker service create --name=cockroachdb2 --network=my_network --mount type=volume,source=cockroach-data2,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.0.1 start --insecure --join=cockroachdb1

# CoachroachDB  -p 26259:26257 -p 3032:8080 
docker service rm cockroachdb3
docker service create --name=cockroachdb3 --network=my_network --mount type=volume,source=cockroach-data3,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.0.1 start --insecure --join=cockroachdb1
