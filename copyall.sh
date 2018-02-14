#!/bin/bash
export GOOS=linux
export CGO_ENABLED=0

cd accountservice;go get;go build -o accountservice-linux-amd64;echo built `pwd`;cd ..
cd vipservice;go get;go build -o vipservice-linux-amd64;echo built `pwd`;cd ..
cd healthchecker;go get;go build -o healthchecker-linux-amd64;echo built `pwd`;cd ..
cd imageservice;go get;go build -o imageservice-linux-amd64;echo built `pwd`;cd ..
cd dataservice;go get;go build -o dataservice-linux-amd64;echo built `pwd`;cd ..


export GOOS=darwin

cp healthchecker/healthchecker-linux-amd64 accountservice/
cp healthchecker/healthchecker-linux-amd64 vipservice/
cp healthchecker/healthchecker-linux-amd64 imageservice/
cp healthchecker/healthchecker-linux-amd64 dataservice/

docker build -t someprefix/accountservice accountservice/
docker service rm accountservice
docker service create --log-driver=gelf --log-opt gelf-address=udp://192.168.99.100:12202 --log-opt gelf-compression-type=none --name=accountservice --replicas=1 --network=my_network -p=6767:6767 someprefix/accountservice

docker build -t someprefix/vipservice vipservice/
docker service rm vipservice
docker service create --log-driver=gelf --log-opt gelf-address=udp://192.168.99.100:12202 --log-opt gelf-compression-type=none --name=vipservice --replicas=1 --network=my_network -p=6868:6868 someprefix/vipservice

docker build -t someprefix/imageservice imageservice/
docker service rm imageservice
docker service create --log-driver=gelf --log-opt gelf-address=udp://192.168.99.100:12202 --log-opt gelf-compression-type=none --name=imageservice --replicas=1 --network=my_network -p=7777:7777 someprefix/imageservice

docker build -t someprefix/dataservice dataservice/
docker service rm dataservice
docker service create --log-driver=gelf --log-opt gelf-address=udp://192.168.99.100:12202 --log-opt gelf-compression-type=none --name=dataservice --replicas=1 --network=my_network -p=7070:7070 someprefix/dataservice
