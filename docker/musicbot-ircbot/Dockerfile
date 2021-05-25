FROM alpine:edge
MAINTAINER Lex van Roon <r3boot@r3blog.nl>

RUN apk update \
    && apk upgrade \
    && apk add ca-certificates youtube-dl ffmpeg \
    && addgroup -S -g 836 git \
    && adduser -S -G git -u 836 git -s /bin/sh

COPY files/run_musicbot /run_musicbot
COPY musicbot/files/musicbot /usr/bin/musicbot

ENTRYPOINT ["/run_musicbot"]
CMD [""]
