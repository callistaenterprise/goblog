#!/bin/bash
export GOOS=linux
export CGO_ENABLED=0

cd accountservice;go get;go build -o accountservice-linux-amd64;echo built `pwd`;cd ..
cd vipservice;go get;go build -o vipservice-linux-amd64;echo built `pwd`;cd ..
cd healthchecker;go get;go build -o healthchecker-linux-amd64;echo built `pwd`;cd ..
cd imageservice;go get;go build -o imageservice-linux-amd64;echo built `pwd`;cd ..

export GOOS=darwin

cp healthchecker/healthchecker-linux-amd64 accountservice/
cp healthchecker/healthchecker-linux-amd64 vipservice/
cp healthchecker/healthchecker-linux-amd64 imageservice/

docker build -t eriklupander/accountservice accountservice/

docker build -t eriklupander/vipservice vipservice/

docker build -t eriklupander/imageservice imageservice/

docker push eriklupander/accountservice:latest
docker push eriklupander/vipservice:latest
docker push eriklupander/imageservice:latest