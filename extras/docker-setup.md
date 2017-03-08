## 1. Create a Swarm Manager

    docker-machine create \
      --driver virtualbox \
      --virtualbox-cpu-count 2 \
      --virtualbox-memory 2048 \
      --virtualbox-disk-size 20000 \
      swarm-manager-1
      
      docker $(docker-machine config swarm-manager-1) swarm init --advertise-addr $(docker-machine ip swarm-manager-1)

These commands created a swarm manager instance for us running on Virtualbox with 2 CPUs and 2 GB of RAM and then made it a Docker Swarm manager node.

## 2. Direct docker commands to a manager in the cluster:

    eval $(docker-machine env swarm-manager-1)
    
## 3. Verify that the cluster is running
      docker-machine ls
      
      NAME              ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER    ERRORS
      default           -        virtualbox   Stopped                                       Unknown   
      swarm-manager-1   *        virtualbox   Running   tcp://192.168.99.100:2376           v1.13.0   

## 4. Store the IP address of the Swarm Manager node in an Env-var

     ManagerIP=`docker-machine ip swarm-manager-1`
    
## 5. Verify using a Docker visualizer

    docker service create \
      --name=viz \
      --publish=8000:8080/tcp \
      --constraint=node.role==manager \
      --mount=type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock \
      manomarks/visualizer
     
This step deploys the third-party component [manomarks/visualizer](https://github.com/ManoMarks/docker-swarm-visualizer) which can be used to show deployed services on the Docker Swarm cluster. After running this command, you should be able to point your web browser to http://192.168.1.100:8000 (you may need to replace the IP address with whatever output you got from the _docker-machine ls_ command)

It should look something like this:

![Viz](/assets/blogg/goblog/part1-viz.png)
