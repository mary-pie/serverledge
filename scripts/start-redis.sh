#!/bin/sh
# grep output == 0 se trova qualcosa
if ! docker network ls | grep redis-net; then
    docker network create --subnet 172.18.0.0/16 redis-net
fi
docker run -d --rm --name Redis-server \
    --network redis-net \
    --ip 172.18.0.2 \
    --publish 6379:6379 \
     redis/redis-stack-server:latest