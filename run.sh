#!/bin/bash 
# this will pull up the containers, and then wait for the interrupt siganls 
# interrupt signal is received will percolate the same thru to the golang container 

_interrupt(){
    echo 'received interrupt'
    docker-compose --env-file dev.env down
}


trap _interrupt SIGTERM
docker-compose --env-file dev.env build && docker-compose --env-file dev.env up
child=$!
wait $child
echo "closing down botmincock"