#!/bin/bash

ID=$(docker ps -a | awk '/postgres:alpine/{ print $1 }' | head -1)

docker exec -it ${ID} psql -U postgres -c "CREATE ROLE musicbot WITH ENCRYPTED PASSWORD 'musicbot';"
docker exec -it ${ID} psql -U postgres -c "ALTER ROLE musicbot WITH LOGIN;"
docker exec -it ${ID} psql -U postgres -c "CREATE DATABASE musicbot OWNER musicbot;"
docker exec -it ${ID} psql -U postgres -c "CREATE EXTENSION pg_trgm;" musicbot
