#!/usr/bin/env bash

PLAYLIST_LOG='/var/log/icecast/playlist.log'

TEMPLATE='/usr/local/share/musicbot/templates/index.html'
OUTPUT='/usr/local/share/musicbot/index.html'

RADIO_2600NL_OGG_LISTENERS=$(grep '/2600nl.ogg' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_2600NL_MP3_LISTENERS=$(grep '/2600nl.mp3' ${PLAYLIST_LOG} | grep -v ^Binary | tail -1 | cut -d\| -f3)
RADIO_2600NL_LISTENERS=$(expr ${RADIO_2600NL_OGG_LISTENERS} + ${RADIO_2600NL_MP3_LISTENERS})
RADIO_2600NL_NOWPLAYING="$(grep '/2600nl.ogg' ${PLAYLIST_LOG} | tail -1 | cut -d\| -f4)"

cat "${TEMPLATE}" | sed \
  -e "s,%2600NL_NOWPLAYING%,${RADIO_2600NL_NOWPLAYING},g" \
  -e "s,%2600NL_NUMLISTENERS%,${RADIO_2600NL_LISTENERS},g" \
  > ${OUTPUT}.new && mv ${OUTPUT}.new ${OUTPUT}