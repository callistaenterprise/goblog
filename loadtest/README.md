# Load test for the Go blog series

### Usage

    mvn gatling:execute -Dusers=1000 -Dduration=30 -DbaseUrl=http://192.168.99.100:6767
    
### Docker usage

    docker stats $(docker ps | awk '{if(NR>1) print $NF}')
