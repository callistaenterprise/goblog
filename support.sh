#!/bin/bash

# RabbitMQ
docker service rm rabbitmq
docker build -t someprefix/rabbitmq support/rabbitmq/
docker service create --name=rabbitmq --replicas=1 --network=my_network -p 1883:1883 -p 5672:5672 -p 15672:15672 someprefix/rabbitmq
