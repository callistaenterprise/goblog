SHELL 	 := /bin/bash

all: clean build swarm-services

clean:
	$(MAKE) -C accountservice/ clean
	$(MAKE) -C dataservice/ clean
	$(MAKE) -C imageservice/ clean
	$(MAKE) -C vipservice/ clean

build:
	$(MAKE) -C accountservice/ build
	$(MAKE) -C dataservice/ build
	$(MAKE) -C imageservice/ build
	$(MAKE) -C vipservice/ build

fmt:
	$(MAKE) -C accountservice/ fmt
	$(MAKE) -C dataservice/ fmt
	$(MAKE) -C imageservice/ fmt
	$(MAKE) -C vipservice/ fmt
	$(MAKE) -C common/ fmt

network:
	docker network create --driver overlay my_network

deploy:
	docker stack deploy -c docker/docker-stack-dev.yaml goblog

swarm-services:
	docker service rm accountservice || true
	docker service create -d --name=accountservice --replicas=1 --network=my_network -p=6767:6767 someprefix/accountservice
	docker service rm vipservice || true
	docker service create -d --name=vipservice --replicas=1 --network=my_network -p=6868:6868 someprefix/vipservice
	docker service rm imageservice || true
	docker service create -d --name=imageservice --replicas=1 --network=my_network -p=7777:7777 someprefix/imageservice
	docker service rm dataservice || true
	docker service create -d --name=dataservice --replicas=1 --network=my_network -p=7070:7070 someprefix/dataservice

rabbitmq:
	docker service rm rabbitmq || true
	docker service create -d --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 rabbitmq:3-management

config-server:
	./support/config-server/gradlew build -p ./support/config-server
	docker build -t someprefix/configserver support/config-server/
	docker service rm configserver || true
	docker service create -d --replicas 1 --name configserver -p 8888:8888 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/configserver

edge-server:
	./support/edge-server/gradlew clean build -p support/edge-server
	docker build -t someprefix/edge-server support/edge-server/
	docker service rm edge-server || true
	docker service create --replicas 1 --name edge-server -p 8765:8765 --network my_network --update-delay 10s --with-registry-auth  --update-parallelism 1 someprefix/edge-server

cockroachdb:
	docker service rm cockroachdb1 || true
	docker service create --name=cockroachdb1 --label cockroachdb --network=my_network -p 26257:26257 -p 3030:8080 --mount type=volume,source=cockroach-data1,target=/cockroach/cockroach-data cockroachdb/cockroach:v19.1.2 start --insecure
	docker service rm cockroachdb2 || true
	docker service create --name=cockroachdb2 --label cockroachdb --network=my_network --mount type=volume,source=cockroach-data2,target=/cockroach/cockroach-data cockroachdb/cockroach:v19.1.2 start --insecure --join=cockroachdb1
	docker service rm cockroachdb3 || true
	docker service create --name=cockroachdb3 --label cockroachdb --network=my_network --mount type=volume,source=cockroach-data3,target=/cockroach/cockroach-data cockroachdb/cockroach:v19.1.2 start --insecure --join=cockroachdb1
