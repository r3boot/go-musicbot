FROM alpine:edge
MAINTAINER Lex van Roon <r3boot@r3blog.nl>

RUN apk update \
    && apk upgrade \
    && apk add ca-certificates youtube-dl ffmpeg \
    && addgroup -S -g 840 musicbot \
    && adduser -S -G musicbot -u 840 musicbot -s /bin/sh

COPY files/run_musicbot /run_musicbot
COPY files/musicbot /usr/bin/musicbot

USER musicbot

ENTRYPOINT ["/run_musicbot"]
CMD [""]
