#!/bin/bash
export GOOS=linux
export CGO_ENABLED=0

cd accountservice;go get;go build -o accountservice-linux-amd64;echo built `pwd`;cd ..
cd vipservice;go get;go build -o vipservice-linux-amd64;echo built `pwd`;cd ..
cd healthchecker;go get;go build -o healthchecker-linux-amd64;echo built `pwd`;cd ..

export GOOS=darwin

cp healthchecker/healthchecker-linux-amd64 accountservice/
cp healthchecker/healthchecker-linux-amd64 vipservice/


docker build -t someprefix/accountservice accountservice/
docker service rm accountservice
docker service create --name=accountservice --replicas=1 --network=my_network -p=6767:6767 someprefix/accountservice

docker build -t someprefix/vipservice vipservice/
docker service rm vipservice
docker service create --name=vipservice --replicas=1 --network=my_network -p=6868:6868 someprefix/vipservice
