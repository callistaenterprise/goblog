#!/bin/bash

docker build -t someprefix/quotes-service quotes-service/
docker service rm quotes-service
docker service create --name=quotes-service --replicas=1 --network=my_network someprefix/quotes-service
