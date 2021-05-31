FROM alpine:edge
MAINTAINER Lex van Roon <r3boot@r3blog.nl>

RUN apk update \
    && apk upgrade \
    && apk add ca-certificates \
    && addgroup -S -g 842 musicbot-ircbot \
    && adduser -S -G musicbot-ircbot -u 842 musicbot-ircbot -s /bin/sh

COPY files/run_musicbot-ircbot /run_musicbot-ircbot
COPY files/musicbot-ircbot /usr/bin/musicbot-ircbot

USER musicbot-ircbot
ENTRYPOINT ["/run_musicbot-ircbot"]
CMD [""]
