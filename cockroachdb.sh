#!/bin/bash

# CoachroachDB master
docker service rm cockroachdb1
docker service create --name=cockroachdb1 --network=my_network -p 26257:26257 -p 3030:8080 --mount type=volume,source=cockroach-data1,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.1.3 start --insecure

# CoachroachDB  -p 26258:26257 -p 3031:8080
docker service rm cockroachdb2
docker service create --name=cockroachdb2 --network=my_network --mount type=volume,source=cockroach-data2,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.1.3 start --insecure --join=cockroachdb1

# CoachroachDB  -p 26259:26257 -p 3032:8080
docker service rm cockroachdb3
docker service create --name=cockroachdb3 --network=my_network --mount type=volume,source=cockroach-data3,target=/cockroach/cockroach-data cockroachdb/cockroach:v1.1.3 start --insecure --join=cockroachdb1
