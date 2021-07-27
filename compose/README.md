# Compose environment
To get the compose environment to work, perform the following steps

## Create required directories
```
cd compose
mkdir -p ./data/{cache,liquidsoap,postgresql}
```

## initialize database
### Bring up the environment:
```
docker-compose up
```
### Figure out the ID of the running container and initialize the db
```
ID=$(docker ps -a | awk '/compose_postgres/{ print $1 }')
docker exec -it ${ID} su - postgres
createuser musicbot
createdb -Omusicbot musicbot
psql
ALTER ROLE musicbot WITH ENCRYPTED PASSWORD 'musicbot';
\c musicbot
CREATE EXTENSION pg_trgm;
```

## Copy audio files
To test the import of existing (old) musicbot mp3's which contain the metadata in ID3 tags, copy the mp3's to the ./data/liquidsoap folder. This is not needed if you upload your own tracks via !dj+

## Interfacing with the bot
The IRC bot will connect to irc.smurfnet.ch/#musicbot-test as 'TestBot'. The webui is reachable as http://localhost:8768/. Do note, to play audio via the webui it is nescecary to modify webui_assets/html/index.html#L21 and webui_assets/js/player.js#L23. In production this whole setup is behind an NGINX proxy, and since the source url is not (yet) configurable, you need to point it to localhost:8000/2600nl.mp3