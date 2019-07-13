# goblog
Code samples for the Go microservice blog series

###
Turbine stream URL: http://192.168.99.100:8282/turbine.stream?cluster=swarm
Hystrix stream URL: http://accountservice:8181/hystrix.stream?cluster=swarm

Make sure Turbine doesn't crash on startup due to AMQ connection problem.

### Setting up Docker Swarm cluster

    docker-machine create --driver virtualbox --virtualbox-cpu-count 4 --virtualbox-memory 6000 --virtualbox-disk-size 30000 swarm-manager-0
    eval "$(docker-machine env swarm-manager-0)"
    docker network create --driver overlay my_network
    docker swarm init --advertise-addr 192.168.99.100
        
##### Adding a worker node


    docker swarm join --token SWMTKN-1-5njothki0tww7gestuh309qgrnr6r357phlsn7ue0r8qmlqnla-181tl1rfou16vv3e7nxrk4ra3 192.168.99.100:2377

    
    
### Deploy spring cloud services

From /goblog

    ./springcloud.sh
    ./support.sh
    
### Deploy microservices

    ./copyall.sh

### Running the CockroachDB client
Find a container running the _cockroachdb/cockroach_ container using _docker ps_ and note the container ID. Then we'll use _docker exec_ to launch the SQL CLI:  
   
    > docker exec -it 10f4b6c727f8 ./cockroach sql --insecure
