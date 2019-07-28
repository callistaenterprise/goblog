# goblog
Code samples for the Go microservice blog series

## Changelog
- 2019-07-27: Total rewrite of Go code, including introducing Makefiles for build and a docker-compose file for deployment on Docker Swarm.

### Some URLs to remember
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
   
### Building
You need Go 1.12 or later installed on your system to build from source.

Builds are performed using Makefile(s). In the root _/goblog_ folder:

    make build
    
One can also run all tests, format code using makefile targets.

### Deploying

Deploy using:
 
    make deploy
  
The make target uses _docker stack deploy_ behind the scenes.

## Nice to haves

Here's some minor stuff worth remembering.

### Running the CockroachDB client
Find a container running the _cockroachdb/cockroach_ container using _docker ps_ and note the container ID. Then we'll use _docker exec_ to launch the SQL CLI:  
   
    > docker exec -it 10f4b6c727f8 ./cockroach sql --insecure

### Create user and database for Cockroach

_(this is possibly broken, from part 16 onwards the initial DB setup is performed using the docker stack)_

Originally from: https://github.com/cockroachdb/cockroach/issues/19826#issuecomment-358360851

    DROP USER IF EXISTS cockroach; \
    DROP DATABASE IF EXISTS account CASCADE; \
    CREATE DATABASE IF NOT EXISTS account; \
    CREATE USER cockroach WITH PASSWORD 'password'; \
    GRANT ALL ON DATABASE account TO cockroach;