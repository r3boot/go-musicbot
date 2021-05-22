#!/bin/bash

ID=$(docker ps -a | awk '/postgres:alpine/{ print $1 }')
docker stop ${ID}
sleep 5
docker rm ${ID}

pkill mpd