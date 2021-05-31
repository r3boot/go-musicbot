FROM alpine:edge
MAINTAINER Lex van Roon <r3boot@r3blog.nl>

RUN apk update \
    && apk upgrade \
    && apk add ca-certificates \
    && addgroup -S -g 841 musicbot-webui \
    && adduser -S -G musicbot-webui -u 841 musicbot-webui -s /bin/sh

COPY files/run_musicbot-webui /run_musicbot-webui
COPY files/musicbot-webui /usr/bin/musicbot-webui

USER musicbot-webui
ENTRYPOINT ["/run_musicbot-webui"]
CMD [""]
