# goblog
Code samples for the Go microservice blog series

###
Turbine stream URL: http://192.168.99.100:8282/turbine.stream?cluster=swarm
Hystrix stream URL: http://accountservice:8181/hystrix.stream?cluster=swarm

Make sure Turbine doesn't crash on startup due to AMQ connection problem.

### Setting up Docker Swarm cluster

    docker-machine create --driver virtualbox --virtualbox-cpu-count 2 --virtualbox-memory 2048 --virtualbox-disk-size 20000 swarm-manager-0
    eval "$(docker-machine env swarm-manager-0)"
    docker network create --driver overlay my_network
    docker swarm init --advertise-addr 192.168.99.100
    
    
### Deploy spring cloud services

From /goblog

    ./springcloud.sh
    ./support.sh
    
### Deploy microservices

    ./copyall.sh
