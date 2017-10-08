#!/bin/bash

export GOOS=linux
export CGO_ENABLED=0

cd go-turbine;go get;go build -o go-turbine-linux-amd64;echo built `pwd`;cd ..

export GOOS=darwin

docker build -t eriklupander/go-turbine go-turbine/
docker service rm go-turbine
docker service create --name=go-turbine -p=8383:8383 --replicas=1 --network=my_network eriklupander/go-turbine
