#!/bin/sh

docker run --rm --network host -v $(pwd)/data/postgres:/var/lib/postgresql/data -e POSTGRES_PASSWORD=password postgres:alpine > ./build/postgres.log 2>&1 &
mpd --no-daemon ./config/mpd.conf > ./build/mpd.log 2>&1 &