#!/bin/bash

docker service rm quotes-service
docker service create --replicas 1 --name quotes-service -p 8080:8080 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 eriklupander/quotes-service
