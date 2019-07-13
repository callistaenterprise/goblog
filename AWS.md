# Setting up cluster on AWS

_Remember that you must push all images declared in docker-stack.yml to docker hub._

1. Use Cloudformation Docker Swarm mode template to set up environment. I've used 1 manager and 1 worker. I've named my env "DockerSwarmGoDemo".

2. Make sure you can SSH into the created manager node. May need to disable source check in EC2 Admin GUI and chmod 400 KEYFILE and/or open up Inbound rules in the ELBs.

3. SSH into manager node. E.g. ssh -i ~/.ssh/MYKEY.pem docker@34.240.204.170

4. Run:

    docker swarm init
    
Make sure you save the TOKEN stuff such as:

    docker swarm join --token SWMTKN-1-2x3u57q7gs8s3or6on1cz4i3t21hf3o21dse7dbt71yhz0oryg-2bzpoa4n2694notj8cubjq3li 172.31.11.84:2377
    
5. SSH into the other node and run:

    docker swarm join --token SWMTKN-1-2x3u57q7gs8s3or6on1cz4i3t21hf3o21dse7dbt71yhz0oryg-2bzpoa4n2694notj8cubjq3li 172.31.11.84:2377
    
  

6. Upload stack file:

    scp -i ~/.ssh/MYKEY.pem docker-stack.yml docker@34.240.204.170:/home/docker
   
7. Deploy stack in manager node:
                
    docker stack deploy -c docker-stack.yml DockerSwarmGoDemo
    Creating network DockerSwarmGoDemo_frontend
    Creating network DockerSwarmGoDemo_default
    Creating network DockerSwarmGoDemo_backend
    Creating service DockerSwarmGoDemo_rabbitmq
    Creating service DockerSwarmGoDemo_configserver
    Creating service DockerSwarmGoDemo_edge-server
    Creating service DockerSwarmGoDemo_zipkin
    Creating service DockerSwarmGoDemo_accountservice
    Creating service DockerSwarmGoDemo_vipservice
    Creating service DockerSwarmGoDemo_imageservice
    Creating service DockerSwarmGoDemo_dvizz


